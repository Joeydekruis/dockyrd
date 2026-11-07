package docker

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
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
	NanoCPUs     int64
	PidsLimit    *int64
	AttachEdge   bool
	SkipProjectNetwork bool
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
	if in.AttachEdge {
		if err := c.EnsureEdgeNetwork(ctx); err != nil {
			return "", err
		}
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
	if in.NanoCPUs > 0 {
		hostConfig.Resources.NanoCPUs = in.NanoCPUs
	}
	if in.PidsLimit != nil && *in.PidsLimit > 0 {
		hostConfig.Resources.PidsLimit = in.PidsLimit
	}
	applyUserContainerHardening(hostConfig, in.MemLimitMB, false)
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
	if !in.SkipProjectNetwork {
		if err := c.moby.NetworkConnect(ctx, projectNet, resp.ID, nil); err != nil {
			_ = c.moby.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
			return "", err
		}
	}
	if in.AttachEdge {
		if err := c.moby.NetworkConnect(ctx, EdgeNetworkName, resp.ID, nil); err != nil {
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

// WaitAppHealthy waits for TCP on port, then optionally HTTP GET healthPath (2xx/3xx).
func (c *Client) WaitAppHealthy(ctx context.Context, containerID string, port int, healthPath string, timeout time.Duration) error {
	started := time.Now()
	if err := c.WaitTCPHealthy(ctx, containerID, port, timeout); err != nil {
		return err
	}
	path := strings.TrimSpace(healthPath)
	if path == "" {
		return nil
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	left := timeout - time.Since(started)
	if left < 5*time.Second {
		left = 5 * time.Second
	}
	deadline := time.Now().Add(left)
	client := &http.Client{Timeout: 5 * time.Second}
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
		ip := containerIP(inspect)
		if ip == "" {
			time.Sleep(2 * time.Second)
			continue
		}
		url := fmt.Sprintf("http://%s%s", net.JoinHostPort(ip, strconv.Itoa(port)), path)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 400 {
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("http health check timed out for path %q", path)
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

// FindComposeServiceContainer returns a running container for a compose service name.
func (c *Client) FindComposeServiceContainer(ctx context.Context, serviceID, deploymentID, composeService string) (string, error) {
	list, err := c.ListManaged(ctx)
	if err != nil {
		return "", err
	}
	composeService = strings.TrimSpace(composeService)
	for _, item := range list {
		if item.Identity.ServiceID != serviceID || item.Identity.DeploymentID != deploymentID || item.State != "running" {
			continue
		}
		if item.Identity.ComposeService == composeService {
			return item.ID, nil
		}
	}
	return "", fmt.Errorf("compose container %q not found", composeService)
}
