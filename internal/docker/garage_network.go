package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
)

// EnsureGarageOnProjectNetwork connects dy-garage to a project network when buckets exist.
func (c *Client) EnsureGarageOnProjectNetwork(ctx context.Context, projectSlug string) error {
	if err := c.EnsureProjectNetwork(ctx, projectSlug); err != nil {
		return err
	}
	garageID, err := c.findContainerByName(ctx, garageContainerName)
	if err != nil || garageID == "" {
		return nil
	}
	projectNet := ProjectNetworkName(projectSlug)
	return c.moby.NetworkConnect(ctx, projectNet, garageID, nil)
}

func (c *Client) findContainerByName(ctx context.Context, name string) (string, error) {
	list, err := c.ListManaged(ctx)
	if err != nil {
		return "", err
	}
	for _, item := range list {
		if item.Name == name {
			return item.ID, nil
		}
	}
	// fallback: inspect by exact name
	listAll, err := c.moby.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", err
	}
	for _, item := range listAll {
		for _, n := range item.Names {
			if n == "/"+name || n == name {
				return item.ID, nil
			}
		}
	}
	return "", fmt.Errorf("container %q not found", name)
}
