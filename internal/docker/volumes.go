package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/errdefs"
)

// RemoveVolume deletes a named Docker volume. Idempotent when the volume is already gone.
func (c *Client) RemoveVolume(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	err := c.moby.VolumeRemove(ctx, name, true)
	if err != nil && errdefs.IsNotFound(err) {
		return nil
	}
	return err
}
