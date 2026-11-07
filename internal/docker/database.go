package docker

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
)

// RunDatabaseInput configures a postgres or redis container on the project network only.
type RunDatabaseInput struct {
	Name         string
	Image        string
	Env          []string
	Identity     Identity
	ProjectSlug  string
	ServiceSlug  string
	MemLimitMB   int
	PublicPort   int
	InternalPort int
	DataMount    string
	Cmd          []string
	SkipProjectNetwork bool
}

// EnsureVolume creates a named volume when missing.
func (c *Client) EnsureVolume(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("volume name required")
	}
	_, err := c.moby.VolumeInspect(ctx, name)
	if err == nil {
		return nil
	}
	_, err = c.moby.VolumeCreate(ctx, volume.CreateOptions{
		Name: name,
		Labels: map[string]string{
			LabelManaged: managedTrue,
			LabelKind:    KindSystem,
		},
	})
	return err
}

// RunDatabase creates and starts a database container with a persistent volume.
func (c *Client) RunDatabase(ctx context.Context, in RunDatabaseInput) (string, error) {
	if in.MemLimitMB <= 0 {
		in.MemLimitMB = 512
	}
	if in.InternalPort <= 0 {
		in.InternalPort = 5432
	}
	if err := c.EnsureProjectNetwork(ctx, in.ProjectSlug); err != nil {
		return "", err
	}
	volName := VolumeName(in.ServiceSlug)
	if err := c.EnsureVolume(ctx, volName); err != nil {
		return "", err
	}

	portKey := nat.Port(fmt.Sprintf("%d/tcp", in.InternalPort))
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		Resources: container.Resources{
			Memory: int64(in.MemLimitMB) * 1024 * 1024,
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: volName,
				Target: in.DataMount,
			},
		},
	}
	if in.PublicPort > 0 {
		hostConfig.PortBindings = nat.PortMap{
			portKey: []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: strconv.Itoa(in.PublicPort)}},
		}
	}
	applyUserContainerHardening(hostConfig, in.MemLimitMB, true)
	applyAppsCgroup(hostConfig)

	labels := in.Identity.Labels()
	exposed := nat.PortSet{portKey: struct{}{}}
	resp, err := c.moby.ContainerCreate(ctx, &container.Config{
		Image:        in.Image,
		Env:          in.Env,
		Cmd:          in.Cmd,
		Labels:       labels,
		ExposedPorts: exposed,
	}, hostConfig, nil, nil, in.Name)
	if err != nil {
		return "", err
	}

	projectNet := ProjectNetworkName(in.ProjectSlug)
	if !in.SkipProjectNetwork {
		if err := c.moby.NetworkConnect(ctx, projectNet, resp.ID, &network.EndpointSettings{
			Aliases: []string{in.ServiceSlug},
		}); err != nil {
			_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
			return "", err
		}
	}
	if err := c.moby.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return "", err
	}
	return resp.ID, nil
}
