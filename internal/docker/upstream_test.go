package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

func TestContainerIPOnNetworkPrefersEdge(t *testing.T) {
	inspect := container.InspectResponse{
		NetworkSettings: &container.NetworkSettings{
			Networks: map[string]*network.EndpointSettings{
				"dy-project-wp": {IPAddress: "172.18.0.2"},
				EdgeNetworkName: {IPAddress: "10.200.0.5"},
			},
		},
	}
	got := containerIPOnNetwork(inspect, EdgeNetworkName)
	if got != "10.200.0.5" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveHostUpstreamDialNilClient(t *testing.T) {
	var c *Client
	if got := c.ResolveHostUpstreamDial(t.Context(), "dy-api-01abc:3000"); got != "dy-api-01abc:3000" {
		t.Fatalf("got %q", got)
	}
}
