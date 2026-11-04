package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
)

// CopyFileToContainer uploads a single file into destDir inside the container.
func (c *Client) CopyFileToContainer(ctx context.Context, containerID, destDir, fileName string, content io.Reader, size int64) error {
	if c == nil {
		return nil
	}
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := tw.WriteHeader(&tar.Header{
		Name: fileName,
		Mode: 0o644,
		Size: size,
	}); err != nil {
		return err
	}
	if _, err := io.Copy(tw, content); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return c.moby.CopyToContainer(ctx, containerID, destDir, &buf, container.CopyToContainerOptions{})
}
