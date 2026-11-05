package docker

import "context"

// FindContainerByName returns the container ID for an exact name (e.g. dy-garage).
func (c *Client) FindContainerByName(ctx context.Context, name string) (string, error) {
	if c == nil {
		return "", nil
	}
	return c.findContainerByName(ctx, name)
}
