package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
)

// HostPTY is an interactive root shell on the panel host.
type HostPTY struct {
	ptmx *os.File
	cmd  *exec.Cmd
}

// StartHostShell starts a login shell in a new PTY on the host.
func StartHostShell(ctx context.Context) (*HostPTY, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}
	cmd := exec.CommandContext(ctx, shell, "-l")
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("host pty: %w", err)
	}
	return &HostPTY{ptmx: ptmx, cmd: cmd}, nil
}

func (h *HostPTY) Read(p []byte) (int, error) {
	if h == nil || h.ptmx == nil {
		return 0, io.EOF
	}
	return h.ptmx.Read(p)
}

func (h *HostPTY) Write(p []byte) (int, error) {
	if h == nil || h.ptmx == nil {
		return 0, io.EOF
	}
	return h.ptmx.Write(p)
}

// Resize sets the PTY window size.
func (h *HostPTY) Resize(cols, rows uint) error {
	if h == nil || h.ptmx == nil || cols == 0 || rows == 0 {
		return nil
	}
	return pty.Setsize(h.ptmx, &pty.Winsize{Cols: uint16(cols), Rows: uint16(rows)})
}

// Close ends the shell session.
func (h *HostPTY) Close() error {
	if h == nil {
		return nil
	}
	if h.ptmx != nil {
		_ = h.ptmx.Close()
	}
	if h.cmd != nil && h.cmd.Process != nil {
		_ = h.cmd.Process.Kill()
		_, _ = h.cmd.Process.Wait()
	}
	return nil
}
