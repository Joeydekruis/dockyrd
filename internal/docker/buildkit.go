package docker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/container"
)

const (
	BuildkitContainerName = "dy-buildkit"
	BuildkitImage         = "moby/buildkit:buildx-stable-1"
)

// EnsureBuildkit starts the system BuildKit container and ensures a buildx builder exists.
func (c *Client) EnsureBuildkit(ctx context.Context) error {
	if c == nil || c.moby == nil {
		return errors.New("docker client required")
	}
	inspect, err := c.moby.ContainerInspect(ctx, BuildkitContainerName)
	if err != nil {
		_, err = c.moby.ContainerCreate(ctx, &container.Config{
			Image: BuildkitImage,
			Labels: map[string]string{
				LabelManaged: managedTrue,
				LabelKind:    KindSystem,
			},
		}, &container.HostConfig{
			Privileged: true,
		}, nil, nil, BuildkitContainerName)
		if err != nil {
			return err
		}
		if err := c.moby.ContainerStart(ctx, BuildkitContainerName, container.StartOptions{}); err != nil {
			return err
		}
	} else if !inspect.State.Running {
		if err := c.moby.ContainerStart(ctx, BuildkitContainerName, container.StartOptions{}); err != nil {
			return err
		}
	}
	return ensureBuildxBuilder(ctx)
}

func ensureBuildxBuilder(ctx context.Context) error {
	name := strings.TrimSpace(os.Getenv("DOCKYARD_BUILDX_BUILDER"))
	if name == "" {
		name = BuildkitContainerName
	}
	check := exec.CommandContext(ctx, "docker", "buildx", "inspect", name)
	if check.Run() == nil {
		return nil
	}
	create := exec.CommandContext(ctx, "docker", "buildx", "create",
		"--name", name,
		"--driver", "docker-container",
		BuildkitContainerName,
	)
	return create.Run()
}
