package docker

import (
	"context"
)

// DockerRootDir returns the daemon's data root (e.g. /var/lib/docker).
func (c *Client) DockerRootDir(ctx context.Context) (string, error) {
	if c == nil || c.moby == nil {
		return "", nil
	}
	info, err := c.moby.Info(ctx)
	if err != nil {
		return "", err
	}
	return info.DockerRootDir, nil
}
