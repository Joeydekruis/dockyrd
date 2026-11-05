package docker

import (
	"context"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

const garageContainerName = "dy-garage"

// RunGarage starts or recreates the system Garage S3 container on dy-edge.
func (c *Client) RunGarage(ctx context.Context, cfg config.Config, _ string) error {
	if err := c.EnsureEdgeNetwork(ctx); err != nil {
		return err
	}
	list, err := c.ListManaged(ctx)
	if err != nil {
		return err
	}
	for _, item := range list {
		if item.Name == garageContainerName {
			if item.State == "running" {
				return nil
			}
			_ = c.StopAndRemove(ctx, item.ID, 10)
			break
		}
	}
	if err := c.PullImage(ctx, "dxflrs/garage:v1.2.0"); err != nil {
		return err
	}
	configPath := filepath.Join(cfg.ConfigDir, "garage", "garage.toml")
	dataRoot := filepath.Join(cfg.DataDir, "garage")
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		Mounts: []mount.Mount{
			{Type: mount.TypeBind, Source: configPath, Target: "/etc/garage.toml", ReadOnly: true},
			{Type: mount.TypeBind, Source: filepath.Join(dataRoot, "meta"), Target: "/var/lib/garage/meta"},
			{Type: mount.TypeBind, Source: filepath.Join(dataRoot, "data"), Target: "/var/lib/garage/data"},
		},
	}
	applyAppsCgroup(hostConfig)
	identity := Identity{Kind: KindSystem, ServerID: "local"}
	resp, err := c.moby.ContainerCreate(ctx, &container.Config{
		Image:  "dxflrs/garage:v1.2.0",
		Labels: identity.Labels(),
		Cmd:    []string{"-c", "/etc/garage.toml", "server"},
	}, hostConfig, nil, nil, garageContainerName)
	if err != nil {
		return err
	}
	if err := c.moby.NetworkConnect(ctx, EdgeNetworkName, resp.ID, nil); err != nil {
		_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return err
	}
	return c.moby.ContainerStart(ctx, resp.ID, container.StartOptions{})
}
