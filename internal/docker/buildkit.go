package docker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/errdefs"
)

const (
	BuildkitContainerName = "dy-buildkit"
	BuildkitImage         = "moby/buildkit:buildx-stable-1"
)

// EnsureBuildkit starts the system BuildKit container and ensures a buildx builder exists.
func (c *Client) EnsureBuildkit(ctx context.Context, dataDir string) error {
	if c == nil || c.moby == nil {
		return errors.New("docker client required")
	}
	cfgPath, err := RenderBuildkitToml(dataDir)
	if err != nil {
		return err
	}
	if err := c.EnsureVolume(ctx, buildkitDataVolume); err != nil {
		return err
	}
	cfgDir := filepath.Dir(cfgPath)
	inspect, err := c.moby.ContainerInspect(ctx, BuildkitContainerName)
	if err != nil && !errdefs.IsNotFound(err) {
		return err
	}
	hostConfig := &container.HostConfig{
		Privileged:    true,
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		Mounts: []mount.Mount{
			{Type: mount.TypeVolume, Source: buildkitDataVolume, Target: "/var/lib/buildkit"},
			{Type: mount.TypeBind, Source: cfgDir, Target: "/etc/buildkit", ReadOnly: true},
		},
	}
	applyBuildCgroup(hostConfig)
	if errdefs.IsNotFound(err) {
		_, err = c.moby.ContainerCreate(ctx, &container.Config{
			Image: BuildkitImage,
			Cmd: []string{
				"buildkitd",
				"--config", "/etc/buildkit/buildkitd.toml",
			},
			Labels: map[string]string{
				LabelManaged: managedTrue,
				LabelKind:    KindSystem,
			},
		}, hostConfig, nil, nil, BuildkitContainerName)
		if err != nil {
			return err
		}
		if err := c.moby.ContainerStart(ctx, BuildkitContainerName, container.StartOptions{}); err != nil {
			return err
		}
	} else {
		if !inspect.State.Running {
			if err := c.moby.ContainerStart(ctx, BuildkitContainerName, container.StartOptions{}); err != nil {
				return err
			}
		}
	}
	return ensureBuildxBuilder(ctx)
}

func ensureBuildxBuilder(ctx context.Context) error {
	name := strings.TrimSpace(os.Getenv("dockyrd_BUILDX_BUILDER"))
	if name == "" {
		name = BuildkitContainerName
	}
	check := exec.CommandContext(ctx, "docker", "buildx", "inspect", name)
	check.Env = append(os.Environ(), build.DockerCLIEnv()...)
	if check.Run() == nil {
		return nil
	}
	create := exec.CommandContext(ctx, "docker", "buildx", "create",
		"--name", name,
		"--driver", "docker-container",
		BuildkitContainerName,
	)
	create.Env = append(os.Environ(), build.DockerCLIEnv()...)
	return create.Run()
}
