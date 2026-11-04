package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
)

// ExecStream runs a command in a container and copies stdout to w (stderr is merged into diagnostics on error).
func (c *Client) ExecStream(ctx context.Context, containerID string, cmd []string, w io.Writer) error {
	if c == nil || containerID == "" {
		return nil
	}
	execID, err := c.moby.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return err
	}
	attach, err := c.moby.ContainerExecAttach(ctx, execID.ID, container.ExecAttachOptions{})
	if err != nil {
		return err
	}
	defer attach.Close()
	if _, err := io.Copy(w, attach.Reader); err != nil {
		return err
	}
	var inspect container.ExecInspect
	inspect, err = c.moby.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return err
	}
	if inspect.ExitCode != 0 {
		return fmt.Errorf("exec exit code %d", inspect.ExitCode)
	}
	return nil
}

// CopyFromContainerPath reads a file from a container filesystem.
func (c *Client) CopyFromContainerPath(ctx context.Context, containerID, path string) (io.ReadCloser, error) {
	reader, _, err := c.moby.CopyFromContainer(ctx, containerID, path)
	return reader, err
}
