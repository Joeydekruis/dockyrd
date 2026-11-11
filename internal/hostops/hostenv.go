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
	return filepath.Join(h.Cfg.ConfigDir, "nftables", "dockyrd.nft")
}

// LegacyNFTRulesPath is the pre-0.1.x rules location (some hosts could not create /etc/nftables.d).
func (h HostEnv) LegacyNFTRulesPath() string {
	if h.HostRoot() != "" {
		return ""
	}
	return "/etc/nftables.d/dockyrd.nft"
}

func (h HostEnv) AptConfDDir() string {
	if root := h.HostRoot(); root != "" {
		return filepath.Join(root, "apt", "apt.conf.d")
	}
	return "/etc/apt/apt.conf.d"
}

func (h HostEnv) UnattendedAptConfPath() string {
	return filepath.Join(h.AptConfDDir(), "50dockyrd-unattended-upgrades")
}

func (h HostEnv) UnattendedAutoUpgradesPath() string {
	return filepath.Join(h.AptConfDDir(), "20dockyrd-auto-upgrades")
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

// ReadNFTRules returns dockyrd nftables rules from the managed path or the legacy system path.
func (h HostEnv) ReadNFTRules() ([]byte, error) {
	data, err := h.ReadFile(h.NFTRulesPath())
	if err == nil {
		return data, nil
	}
	if legacy := h.LegacyNFTRulesPath(); legacy != "" {
		if data, err2 := h.ReadFile(legacy); err2 == nil {
			return data, nil
		}
	}
	return nil, err
}

func (h HostEnv) WriteFile(path string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, data, mode)
}
