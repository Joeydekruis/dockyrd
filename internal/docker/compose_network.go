package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
)

// ConnectProxyToStackNetwork attaches dy-proxy to a compose stack network for upstream DNS.
func (c *Client) ConnectProxyToStackNetwork(ctx context.Context, networkName string) error {
	if c == nil || c.moby == nil {
		return fmt.Errorf("docker client required")
	}
	networkName = strings.TrimSpace(networkName)
	if networkName == "" {
		return fmt.Errorf("network name required")
	}
	list, err := c.moby.NetworkList(ctx, network.ListOptions{
		Filters: filters.NewArgs(filters.Arg("name", networkName)),
	})
	if err != nil {
		return err
	}
	var netID string
	for _, n := range list {
		if n.Name == networkName {
			netID = n.ID
			break
		}
	}
	if netID == "" {
		return fmt.Errorf("network %q not found", networkName)
	}
	inspect, err := c.moby.ContainerInspect(ctx, ProxyContainerName)
	if err != nil {
		return fmt.Errorf("inspect dy-proxy: %w", err)
	}
	for _, n := range inspect.NetworkSettings.Networks {
		if n != nil && n.NetworkID == netID {
			return nil
		}
	}
	err = c.moby.NetworkConnect(ctx, netID, inspect.ID, &network.EndpointSettings{})
	if err != nil && !errdefs.IsConflict(err) {
		return fmt.Errorf("connect dy-proxy to %s: %w", networkName, err)
	}
	return nil
}
