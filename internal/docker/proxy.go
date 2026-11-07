package docker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/errdefs"
)

const (
	ProxyContainerName  = "dy-proxy"
	ProxyImage          = "caddy:2-alpine"
	proxySpecLabel      = "dev.dockyrd.proxy_spec"
	proxySpecHostV1     = "host-v1"
	caddyDirInContainer = "/var/lib/dockyrd/caddy"
)

// EnsureProxy runs the system Caddy container (dy-proxy) with host networking on 80/443.
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
	if err := os.MkdirAll(filepath.Join(caddyDir, "certificates"), 0o750); err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.SitesDir(), 0o755); err != nil {
		return err
	}
	bootstrapPath := filepath.Join(caddyDir, "bootstrap.json")
	raw, err := proxyBootstrapJSON()
	if err != nil {
		return err
	}
	if err := os.WriteFile(bootstrapPath, raw, 0o644); err != nil {
		return err
	}

	if err := c.PullImage(ctx, ProxyImage); err != nil {
		return fmt.Errorf("pull %s: %w", ProxyImage, err)
	}

	inspect, err := c.moby.ContainerInspect(ctx, ProxyContainerName)
	if err == nil {
		spec := inspect.Config.Labels[proxySpecLabel]
		if spec != proxySpecHostV1 {
			if err := c.removeProxyContainer(ctx); err != nil {
				return err
			}
			err = errdefs.NotFound(errors.New("recreate dy-proxy"))
		} else if !inspect.State.Running {
			if err := c.moby.ContainerStart(ctx, ProxyContainerName, container.StartOptions{}); err != nil {
				return fmt.Errorf("start dy-proxy: %w", err)
			}
		}
	}
	if err != nil && !errdefs.IsNotFound(err) {
		return err
	}

	if errdefs.IsNotFound(err) {
		hostCfg := &container.HostConfig{
			NetworkMode:   "host",
			RestartPolicy: container.RestartPolicy{Name: "always"},
			Mounts: []mount.Mount{
				{Type: mount.TypeBind, Source: caddyDir, Target: caddyDirInContainer},
				{Type: mount.TypeBind, Source: cfg.SitesDir(), Target: "/srv/sites", ReadOnly: true},
			},
		}
		applySystemCgroup(hostCfg)
		_, err = c.moby.ContainerCreate(ctx, &container.Config{
			Image: ProxyImage,
			Cmd: []string{
				"caddy", "run",
				"--config", caddyDirInContainer + "/bootstrap.json",
			},
			Labels: map[string]string{
				LabelManaged:   managedTrue,
				LabelKind:      KindSystem,
				proxySpecLabel: proxySpecHostV1,
			},
		}, hostCfg, nil, nil, ProxyContainerName)
		// cgroup applied on create only; existing dy-proxy keeps prior placement until recreated
		if err != nil {
			return fmt.Errorf("create dy-proxy: %w", err)
		}
		if err := c.moby.ContainerStart(ctx, ProxyContainerName, container.StartOptions{}); err != nil {
			return fmt.Errorf("start dy-proxy: %w", err)
		}
	}

	if err := waitForFile(ctx, cfg.CaddyAdminSocket(), 45*time.Second); err != nil {
		logs, _ := c.moby.ContainerLogs(ctx, ProxyContainerName, container.LogsOptions{Tail: "40"})
		if logs != nil {
			defer logs.Close()
			buf := make([]byte, 4096)
			n, _ := logs.Read(buf)
			if n > 0 {
				return fmt.Errorf("%w (dy-proxy logs: %s)", err, strings.TrimSpace(string(buf[:n])))
			}
		}
		return err
	}
	return nil
}

func (c *Client) removeProxyContainer(ctx context.Context) error {
	timeout := 10
	_ = c.moby.ContainerStop(ctx, ProxyContainerName, container.StopOptions{Timeout: &timeout})
	return c.moby.ContainerRemove(ctx, ProxyContainerName, container.RemoveOptions{Force: true})
}

func proxyBootstrapJSON() ([]byte, error) {
	doc := map[string]any{
		"admin": map[string]any{"listen": "unix//" + strings.TrimPrefix(caddyDirInContainer+"/admin.sock", "/")},
		"apps": map[string]any{
			"http": map[string]any{
				"servers": map[string]any{
					"bootstrap": map[string]any{
						"listen":    []string{":80", ":443"},
						"protocols": []string{"h1", "h2"},
						"routes":    []any{},
					},
				},
			},
			"tls": map[string]any{
				"automation": map[string]any{
					"policies": []any{
						map[string]any{
							"issuers": []any{
								map[string]any{
									"module": "acme",
									"email":  "ops@localhost",
								},
							},
						},
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

// ProxyRunning reports whether dy-proxy exists and is running with the current spec.
func (c *Client) ProxyRunning(ctx context.Context) (bool, string) {
	if c == nil || c.moby == nil {
		return false, "docker unavailable"
	}
	inspect, err := c.moby.ContainerInspect(ctx, ProxyContainerName)
	if err != nil {
		return false, "not created"
	}
	if inspect.Config.Labels[proxySpecLabel] != proxySpecHostV1 {
		return false, "outdated spec (restart dockyrd to recreate)"
	}
	if !inspect.State.Running {
		return false, inspect.State.Status
	}
	return true, "running"
}
