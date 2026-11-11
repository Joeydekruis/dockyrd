package docker

import (
	"context"

	"github.com/docker/docker/api/types"
)

// EngineDiskUsage returns Docker disk usage breakdown.
func (c *Client) EngineDiskUsage(ctx context.Context) (types.DiskUsage, error) {
	if c == nil || c.moby == nil {
		return types.DiskUsage{}, ErrUnavailable
	}
	return c.moby.DiskUsage(ctx, types.DiskUsageOptions{})
}
