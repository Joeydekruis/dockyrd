package docker

import (
	"context"
	"net"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
)

// ResolveHostUpstreamDial maps container hostnames to bridge IPs for host-networked dy-proxy.
// Hostnames that are already IPs or unknown containers are returned unchanged.
func (c *Client) ResolveHostUpstreamDial(ctx context.Context, dial string) string {
	if c == nil || c.moby == nil {
		return dial
	}
	dial = strings.TrimSpace(dial)
	if dial == "" {
		return dial
	}
	host, port, err := net.SplitHostPort(dial)
	if err != nil {
		return dial
	}
	if ip := net.ParseIP(host); ip != nil {
		return dial
	}
	if !strings.HasPrefix(host, "dy-") {
		return dial
	}
	inspect, err := c.moby.ContainerInspect(ctx, host)
	if err != nil {
		return dial
	}
	if !inspect.State.Running {
		return dial
	}
	ip := containerIPOnNetwork(inspect, EdgeNetworkName)
	if ip == "" {
		ip = containerIP(inspect)
	}
	if ip == "" {
		return dial
	}
	return net.JoinHostPort(ip, port)
}

func containerIPOnNetwork(inspect container.InspectResponse, networkName string) string {
	if networkName == "" {
		return ""
	}
	if n, ok := inspect.NetworkSettings.Networks[networkName]; ok && n != nil && n.IPAddress != "" {
		return n.IPAddress
	}
	return ""
}

// ParseDialPort returns the port from host:port dial strings (defaults to 80).
func ParseDialPort(dial string) int {
	_, port, err := net.SplitHostPort(strings.TrimSpace(dial))
	if err != nil {
		return 80
	}
	p, err := strconv.Atoi(port)
	if err != nil || p <= 0 {
		return 80
	}
	return p
}
