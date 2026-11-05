package docker

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/go-connections/nat"
)

// RunAppInput configures a new app container.
type RunAppInput struct {
	Name         string
	Image        string
	Port         int
	Env          []string
	Identity     Identity
	ProjectSlug  string
	MemLimitMB   int
}

// PullImage fetches an image reference into the local daemon.
func (c *Client) PullImage(ctx context.Context, ref string) error {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return fmt.Errorf("image ref required")
	}
	out, err := c.moby.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	_, _ = io.Copy(io.Discard, out)
	return nil
}

// RunApp creates and starts a container on project + edge networks.
func (c *Client) RunApp(ctx context.Context, in RunAppInput) (string, error) {
	if in.Port <= 0 {
		in.Port = 80
	}
	if in.MemLimitMB <= 0 {
		in.MemLimitMB = 512
	}
	if err := c.EnsureEdgeNetwork(ctx); err != nil {
		return "", err
	}
	if err := c.EnsureProjectNetwork(ctx, in.ProjectSlug); err != nil {
		return "", err
	}

	labels := in.Identity.Labels()
	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "unless-stopped"},
		Resources: container.Resources{
			Memory: int64(in.MemLimitMB) * 1024 * 1024,
		},
	}
	applyAppsCgroup(hostConfig)

	portKey := nat.Port(fmt.Sprintf("%d/tcp", in.Port))
	exposed := nat.PortSet{portKey: struct{}{}}

	resp, err := c.moby.ContainerCreate(ctx, &container.Config{
		Image:        in.Image,
		Env:          in.Env,
		ExposedPorts: exposed,
		Labels:       labels,
	}, hostConfig, nil, nil, in.Name)
	if err != nil {
		return "", err
	}

	projectNet := ProjectNetworkName(in.ProjectSlug)
	if err := c.moby.NetworkConnect(ctx, projectNet, resp.ID, nil); err != nil {
		_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return "", err
	}
	if err := c.moby.NetworkConnect(ctx, EdgeNetworkName, resp.ID, nil); err != nil {
		_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return "", err
	}
	if err := c.moby.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return "", err
	}
	return resp.ID, nil
}

// StopAndRemove stops a container gracefully then removes it.
func (c *Client) StopAndRemove(ctx context.Context, containerID string, graceSeconds int) error {
	if graceSeconds <= 0 {
		graceSeconds = 30
	}
	timeout := graceSeconds
	_ = c.moby.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
	return c.moby.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}

// WaitTCPHealthy dials containerIP:port until success or timeout.
func (c *Client) WaitTCPHealthy(ctx context.Context, containerID string, port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		inspect, err := c.moby.ContainerInspect(ctx, containerID)
		if err != nil {
			return err
		}
		if !inspect.State.Running {
			time.Sleep(2 * time.Second)
			continue
		}
		ip := containerIP(inspect)
		if ip == "" {
			time.Sleep(2 * time.Second)
			continue
		}
		addr := net.JoinHostPort(ip, strconv.Itoa(port))
		d := net.Dialer{Timeout: 2 * time.Second}
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("health check timed out after %s", timeout)
}

func containerIP(inspect container.InspectResponse) string {
	for _, net := range inspect.NetworkSettings.Networks {
		if net != nil && net.IPAddress != "" {
			return net.IPAddress
		}
	}
	return ""
}

// ContainerLogs returns stdout+stderr since a timestamp.
func (c *Client) ContainerLogs(ctx context.Context, containerID string, tail string) (string, error) {
	if tail == "" {
		tail = "200"
	}
	out, err := c.moby.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	})
	if err != nil {
		return "", err
	}
	defer out.Close()
	data, err := io.ReadAll(out)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FindLiveContainerForService returns the running container ID for a service deployment label.
func (c *Client) FindLiveContainerForService(ctx context.Context, serviceID, deploymentID string) (string, error) {
	list, err := c.ListManaged(ctx)
	if err != nil {
		return "", err
	}
	for _, item := range list {
		if item.Identity.ServiceID == serviceID && item.Identity.DeploymentID == deploymentID && item.State == "running" {
			return item.ID, nil
		}
	}
	return "", fmt.Errorf("live container not found")
}
