package docker

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

const (
	ProxyContainerName = "dy-proxy"
	ProxyImage         = "caddy:2-alpine"
)

// EnsureProxy runs the system Caddy container (dy-proxy) with admin socket on the data volume.
func (c *Client) EnsureProxy(ctx context.Context, cfg config.Config) error {
	if c == nil || c.moby == nil {
		return errors.New("docker client required")
	}
	if err := c.EnsureEdgeNetwork(ctx); err != nil {
		return err
	}
	caddyDir := cfg.CaddyDir()
	if err := os.MkdirAll(caddyDir, 0o750); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.SitesDir(), 0o750); err != nil {
		return err
	}
	bootstrapPath := filepath.Join(caddyDir, "bootstrap.json")
	raw, err := proxyBootstrapJSON(cfg.CaddyAdminSocket())
	if err != nil {
		return err
	}
	if err := os.WriteFile(bootstrapPath, raw, 0o644); err != nil {
		return err
	}

	inspect, err := c.moby.ContainerInspect(ctx, ProxyContainerName)
	if err != nil {
		_, err = c.moby.ContainerCreate(ctx, &container.Config{
			Image: ProxyImage,
			Cmd: []string{
				"caddy", "run",
				"--config", "/var/lib/dockyard/caddy/bootstrap.json",
			},
			Labels: map[string]string{
				LabelManaged: managedTrue,
				LabelKind:    KindSystem,
			},
		}, &container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port("80/tcp"):  {{HostIP: "0.0.0.0", HostPort: "80"}},
				nat.Port("443/tcp"): {{HostIP: "0.0.0.0", HostPort: "443"}},
			},
			RestartPolicy: container.RestartPolicy{Name: "always"},
			ExtraHosts:    []string{"host.docker.internal:host-gateway"},
			Mounts: []mount.Mount{
				{Type: mount.TypeBind, Source: caddyDir, Target: "/var/lib/dockyard/caddy"},
				{Type: mount.TypeBind, Source: cfg.SitesDir(), Target: "/srv/sites", ReadOnly: true},
			},
		}, &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				EdgeNetworkName: {},
			},
		}, nil, ProxyContainerName)
		if err != nil {
			return err
		}
		if err := c.moby.ContainerStart(ctx, ProxyContainerName, container.StartOptions{}); err != nil {
			return err
		}
	} else if !inspect.State.Running {
		if err := c.moby.ContainerStart(ctx, ProxyContainerName, container.StartOptions{}); err != nil {
			return err
		}
	}
	return waitForFile(ctx, cfg.CaddyAdminSocket(), 30*time.Second)
}

func proxyBootstrapJSON(adminSocket string) ([]byte, error) {
	adminSocket = strings.TrimPrefix(strings.TrimSpace(adminSocket), "/")
	doc := map[string]any{
		"admin": map[string]any{"listen": "unix//" + adminSocket},
		"apps": map[string]any{
			"http": map[string]any{
				"servers": map[string]any{
					"bootstrap": map[string]any{
						"listen": []string{":8080"},
						"routes": []any{},
					},
				},
			},
		},
	}
	return json.MarshalIndent(doc, "", "  ")
}

func waitForFile(ctx context.Context, path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	return errors.New("timed out waiting for caddy admin socket")
}
