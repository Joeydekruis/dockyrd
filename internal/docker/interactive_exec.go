package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/container"
)

// InteractiveExec opens a TTY-attached exec session inside a container.
type InteractiveExec struct {
	conn   io.ReadWriteCloser
	execID string
}

// StartInteractiveExec starts an interactive TTY exec and returns a duplex stream.
func (c *Client) StartInteractiveExec(ctx context.Context, containerID string, cmd []string, execUser string, env []string) (*InteractiveExec, error) {
	if c == nil || containerID == "" {
		return nil, fmt.Errorf("docker unavailable")
	}
	if len(cmd) == 0 {
		cmd = []string{"/bin/sh"}
	}
	opts := container.ExecOptions{
		User:         execUser,
		Cmd:          cmd,
		Env:          env,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	}
	execResp, err := c.moby.ContainerExecCreate(ctx, containerID, opts)
	if err != nil {
		return nil, err
	}
	attach, err := c.moby.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{Tty: true})
	if err != nil {
		return nil, err
	}
	if attach.Conn == nil {
		attach.Close()
		return nil, fmt.Errorf("exec attach returned no connection")
	}
	return &InteractiveExec{conn: attach.Conn, execID: execResp.ID}, nil
}

// Read implements io.Reader.
func (e *InteractiveExec) Read(p []byte) (int, error) {
	if e == nil || e.conn == nil {
		return 0, io.EOF
	}
	return e.conn.Read(p)
}

// Write implements io.Writer.
func (e *InteractiveExec) Write(p []byte) (int, error) {
	if e == nil || e.conn == nil {
		return 0, io.EOF
	}
	return e.conn.Write(p)
}

// ResizeExec informs the container PTY of a new geometry.
func (c *Client) ResizeExec(ctx context.Context, e *InteractiveExec, cols, rows uint) error {
	if c == nil || e == nil || e.execID == "" {
		return nil
	}
	return c.moby.ContainerExecResize(ctx, e.execID, container.ResizeOptions{
		Height: rows,
		Width:  cols,
	})
}

// Close ends the exec attach stream.
func (e *InteractiveExec) Close() error {
	if e == nil || e.conn == nil {
		return nil
	}
	return e.conn.Close()
}
