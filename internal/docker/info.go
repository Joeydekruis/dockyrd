package docker

import (
	"context"
	"strings"
)

// DockerRootDir returns the daemon data root (e.g. /var/lib/docker).
func (c *Client) DockerRootDir(ctx context.Context) (string, error) {
	if c == nil || c.moby == nil {
		return "", nil
	}
	info, err := c.moby.Info(ctx)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(info.DockerRootDir), nil
}
