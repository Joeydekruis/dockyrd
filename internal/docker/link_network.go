package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/network"
)

// EnsureLinkNetwork creates a bridge network for one service link (idempotent).
func (c *Client) EnsureLinkNetwork(ctx context.Context, linkID string) error {
	linkID = strings.TrimSpace(linkID)
	if linkID == "" {
		return nil
	}
	name := LinkNetworkName(linkID)
	_, err := c.moby.NetworkInspect(ctx, name, network.InspectOptions{})
	if err == nil {
		return nil
	}
	_, err = c.moby.NetworkCreate(ctx, name, network.CreateOptions{
		Driver: "bridge",
		Labels: map[string]string{
			LabelManaged: managedTrue,
			LabelKind:    KindSystem,
			LabelLink:    linkID,
		},
	})
	return err
}

// ConnectLinkEndpoint attaches a container to a link network with optional DNS aliases.
func (c *Client) ConnectLinkEndpoint(ctx context.Context, linkID, containerID string, aliases []string) error {
	if containerID == "" {
		return nil
	}
	if err := c.EnsureLinkNetwork(ctx, linkID); err != nil {
		return err
	}
	netName := LinkNetworkName(linkID)
	settings := &network.EndpointSettings{}
	if len(aliases) > 0 {
		settings.Aliases = aliases
	}
	err := c.moby.NetworkConnect(ctx, netName, containerID, settings)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "already") {
		return nil
	}
	return err
}

// EnsureGarageOnLinkNetwork connects dy-garage to a link network (bucket links).
func (c *Client) EnsureGarageOnLinkNetwork(ctx context.Context, linkID string) error {
	garageID, err := c.findContainerByName(ctx, garageContainerName)
	if err != nil || garageID == "" {
		return nil
	}
	return c.ConnectLinkEndpoint(ctx, linkID, garageID, nil)
}
