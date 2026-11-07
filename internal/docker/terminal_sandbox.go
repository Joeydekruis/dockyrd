package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/oklog/ulid/v2"
)

const terminalSandboxImage = "alpine:3.20"

// RunTerminalSandboxInput configures an ephemeral debug container on project/link networks.
type RunTerminalSandboxInput struct {
	Name        string
	ProjectID   string
	ProjectSlug string
	LinksMode   string
	LinkIDs     []string
}

// RunTerminalSandbox starts a hardened throwaway container for the web test shell.
func (c *Client) RunTerminalSandbox(ctx context.Context, in RunTerminalSandboxInput) (string, error) {
	if c == nil {
		return "", fmt.Errorf("docker unavailable")
	}
	in.ProjectSlug = strings.TrimSpace(in.ProjectSlug)
	if in.ProjectSlug == "" {
		return "", fmt.Errorf("project slug required")
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = "dy-term-" + ulid.Make().String()
	}
	if err := c.PullImage(ctx, terminalSandboxImage); err != nil {
		return "", err
	}

	pidsLimit := int64(512)
	memMB := int64(384)
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "no"},
		CapDrop:       []string{"ALL"},
		CapAdd:        []string{"CHOWN", "DAC_OVERRIDE", "FOWNER", "SETGID", "SETUID", "NET_BIND_SERVICE", "NET_RAW"},
		SecurityOpt:   []string{"no-new-privileges:true"},
		Resources: container.Resources{
			Memory:    memMB * 1024 * 1024,
			PidsLimit: &pidsLimit,
		},
		Mounts: []mount.Mount{
			{Type: mount.TypeTmpfs, Target: "/tmp"},
		},
	}
	applyAppsCgroup(hostConfig)

	labels := map[string]string{
		LabelManaged:              managedTrue,
		LabelKind:                 KindSystem,
		LabelProject:              in.ProjectID,
		"dev.dockyard.terminal":   "sandbox",
	}

	startCmd := []string{
		"/bin/sh", "-c",
		"apk add --no-cache --quiet curl bind-tools iputils postgresql-client redis >/dev/null 2>&1; exec sleep infinity",
	}

	resp, err := c.moby.ContainerCreate(ctx, &container.Config{
		Image:  terminalSandboxImage,
		Cmd:    startCmd,
		Labels: labels,
	}, hostConfig, nil, nil, name)
	if err != nil {
		return "", err
	}

	mode := strings.ToLower(strings.TrimSpace(in.LinksMode))
	if mode == "" {
		mode = "strict"
	}
	if mode == "open" {
		if err := c.EnsureProjectNetwork(ctx, in.ProjectSlug); err != nil {
			_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
			return "", err
		}
		projectNet := ProjectNetworkName(in.ProjectSlug)
		if err := c.moby.NetworkConnect(ctx, projectNet, resp.ID, nil); err != nil {
			_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
			return "", err
		}
	} else {
		if len(in.LinkIDs) == 0 {
			_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
			return "", fmt.Errorf("no service links in this project; add a link or use links_mode open")
		}
		for _, linkID := range in.LinkIDs {
			linkID = strings.TrimSpace(linkID)
			if linkID == "" {
				continue
			}
			if err := c.ConnectLinkEndpoint(ctx, linkID, resp.ID, nil); err != nil {
				_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
				return "", err
			}
		}
	}

	if err := c.moby.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return "", err
	}
	return resp.ID, nil
}
