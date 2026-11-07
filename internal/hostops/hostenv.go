package hostops

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

// Commander runs host binaries without shell interpolation.
type Commander interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type execCommander struct{}

func (execCommander) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// HostEnv holds shared paths and command execution for host actions.
type HostEnv struct {
	Cfg config.Config
	Exec Commander
}

func NewHostEnv(cfg config.Config) HostEnv {
	return HostEnv{Cfg: cfg, Exec: execCommander{}}
}

func (h HostEnv) WithExec(c Commander) HostEnv {
	h.Exec = c
	return h
}

// HostRoot returns the prefix for simulated host paths on non-Linux dev hosts.
func (h HostEnv) HostRoot() string {
	if runtime.GOOS == "linux" {
		return ""
	}
	return filepath.Join(h.Cfg.DataDir, "host")
}

func (h HostEnv) SSHDropInDir() string {
	if root := h.HostRoot(); root != "" {
		return filepath.Join(root, "ssh", "sshd_config.d")
	}
	return "/etc/ssh/sshd_config.d"
}

func (h HostEnv) SSHDropInPath() string {
	return filepath.Join(h.SSHDropInDir(), "00-dockyrd.conf")
}

func (h HostEnv) NFTRulesPath() string {
	if root := h.HostRoot(); root != "" {
		return filepath.Join(root, "nftables", "dockyrd.nft")
	}
	return "/etc/nftables.d/dockyrd.nft"
}

func (h HostEnv) UnattendedStampPath() string {
	return filepath.Join(h.Cfg.DataDir, "host", "unattended-upgrades.enabled")
}

func (h HostEnv) OSReleasePath() string {
	if root := h.HostRoot(); root != "" {
		return filepath.Join(root, "etc", "os-release")
	}
	return "/etc/os-release"
}

func (h HostEnv) RebootRequiredPath() string {
	if root := h.HostRoot(); root != "" {
		return filepath.Join(root, "var", "run", "reboot-required")
	}
	return "/var/run/reboot-required"
}

func (h HostEnv) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (h HostEnv) WriteFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, data, mode)
}
