package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

// ValidateCaddyJSON runs caddy validate in a one-off container for the given image.
func (c *Client) ValidateCaddyJSON(ctx context.Context, image string, cfg config.Config, configJSON []byte) error {
	if c == nil || c.moby == nil {
		return fmt.Errorf("docker client required")
	}
	image = strings.TrimSpace(image)
	if image == "" {
		return fmt.Errorf("image required")
	}
	if len(configJSON) == 0 {
		return fmt.Errorf("config required")
	}
	tmpDir, err := os.MkdirTemp("", "dy-proxy-validate-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	cfgPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(cfgPath, configJSON, 0o644); err != nil {
		return err
	}
	hostCfg := &container.HostConfig{
		Mounts: []mount.Mount{
			{Type: mount.TypeBind, Source: cfgPath, Target: "/validate/config.json", ReadOnly: true},
		},
	}
	resp, err := c.moby.ContainerCreate(ctx, &container.Config{
		Image:      image,
		Entrypoint: []string{"caddy"},
		Cmd:        []string{"validate", "--config", "/validate/config.json"},
	}, hostCfg, nil, nil, "")
	if err != nil {
		return fmt.Errorf("create validate container: %w", err)
	}

	defer func() {
		_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
	}()
	if err := c.moby.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("start validate: %w", err)
	}
	statusCh, errCh := c.moby.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	var exitCode int64
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			return err
		}
	case wait := <-statusCh:
		exitCode = wait.StatusCode
	}
	if exitCode != 0 {
		out, _ := c.moby.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: "40"})
		msg := ""
		if out != nil {
			b, _ := io.ReadAll(out)
			_ = out.Close()
			msg = strings.TrimSpace(string(b))
		}
		if msg == "" {
			return fmt.Errorf("caddy validate failed (exit %d)", exitCode)
		}
		return fmt.Errorf("caddy validate failed (exit %d): %s", exitCode, msg)
	}
	return nil
}