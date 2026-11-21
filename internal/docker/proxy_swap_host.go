package docker

import (
	"context"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
)

const proxyRetainName = "dy-proxy-swap-retain"

// WaitForCaddyAdmin blocks until the admin unix socket exists.
func WaitForCaddyAdmin(ctx context.Context, socket string, timeout time.Duration) error {
	return waitForFile(ctx, socket, timeout)
}

// ParkProxyForSwap stops dy-proxy and renames it for rollback.
func (c *Client) ParkProxyForSwap(ctx context.Context) error {
	inspect, err := c.moby.ContainerInspect(ctx, ProxyContainerName)
	if err != nil {
		return err
	}
	if inspect.State.Running {
		timeout := 10
		if err := c.moby.ContainerStop(ctx, ProxyContainerName, container.StopOptions{Timeout: &timeout}); err != nil {
			return err
		}
	}
	return c.moby.ContainerRename(ctx, inspect.ID, proxyRetainName)
}

// RestoreParkedProxy renames the retained container back to dy-proxy and starts it.
func (c *Client) RestoreParkedProxy(ctx context.Context, cfg config.Config) error {
	inspect, err := c.moby.ContainerInspect(ctx, proxyRetainName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return c.EnsureProxy(ctx, cfg)
		}
		return err
	}
	if err := c.moby.ContainerRename(ctx, inspect.ID, ProxyContainerName); err != nil {
		return err
	}
	if !inspect.State.Running {
		return c.moby.ContainerStart(ctx, ProxyContainerName, container.StartOptions{})
	}
	return nil
}

// RemoveParkedProxy deletes the retained container from a successful swap.
func (c *Client) RemoveParkedProxy(ctx context.Context) error {
	timeout := 10
	_ = c.moby.ContainerStop(ctx, proxyRetainName, container.StopOptions{Timeout: &timeout})
	err := c.moby.ContainerRemove(ctx, proxyRetainName, container.RemoveOptions{Force: true})
	if errdefs.IsNotFound(err) {
		return nil
	}
	return err
}

// RecreateProxyWithImage removes dy-proxy and creates it with the given image.
func (c *Client) RecreateProxyWithImage(ctx context.Context, cfg config.Config, image string) error {
	if err := c.RemoveProxyContainer(ctx); err != nil && !errdefs.IsNotFound(err) {
		return err
	}
	return c.ensureProxyImage(ctx, cfg, image)
}

// RemoveProxyContainer stops and removes dy-proxy.
func (c *Client) RemoveProxyContainer(ctx context.Context) error {
	return c.removeProxyContainer(ctx)
}