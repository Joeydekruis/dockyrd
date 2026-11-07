package hostops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

)

const unattendedAptConf = `Unattended-Upgrade::Automatic-Reboot "false";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
`

// UnattendedEnable configures unattended-upgrades on Debian/Ubuntu.
type UnattendedEnable struct {
	Env HostEnv
}

func NewUnattendedEnable(env HostEnv) *UnattendedEnable {
	return &UnattendedEnable{Env: env}
}

func (u *UnattendedEnable) Kind() string { return KindUnattendedEnable }

func (u *UnattendedEnable) Plan(_ context.Context) (Plan, error) {
	distro, err := u.readDistro()
	if err != nil {
		return Plan{
			"summary": "Unattended upgrades not supported on this distro",
			"warn":    err.Error(),
		}, fmt.Errorf("%w: %s", ErrPrecondition, err.Error())
	}
	return Plan{
		"summary": "Enable unattended-upgrades (security only, no auto-reboot)",
		"distro":  distro,
		"stamp":   u.Env.UnattendedStampPath(),
	}, nil
}

func (u *UnattendedEnable) readDistro() (string, error) {
	data, err := u.Env.ReadFile(u.Env.OSReleasePath())
	if err != nil {
		if u.Env.HostRoot() != "" {
			return "debian-sim", nil
		}
		return "", fmt.Errorf("cannot read os-release")
	}
	s := strings.ToLower(string(data))
	if strings.Contains(s, "debian") || strings.Contains(s, "ubuntu") {
		if strings.Contains(s, "ubuntu") {
			return "ubuntu", nil
		}
		return "debian", nil
	}
	return "", fmt.Errorf("only Debian/Ubuntu supported")
}

func (u *UnattendedEnable) Apply(ctx context.Context) (RevertPayload, error) {
	if _, err := u.readDistro(); err != nil {
		return nil, err
	}
	confPath := filepath.Join(u.Env.Cfg.ConfigDir, "unattended-upgrades", "50dockyrd.conf")
	prev := ""
	if data, err := u.Env.ReadFile(confPath); err == nil {
		prev = string(data)
	}
	if err := u.Env.WriteFile(confPath, []byte(unattendedAptConf), 0o644); err != nil {
		return nil, err
	}
	stamp := u.Env.UnattendedStampPath()
	if err := u.Env.WriteFile(stamp, []byte("enabled\n"), 0o600); err != nil {
		return nil, err
	}
	if u.Env.HostRoot() == "" {
		if out, err := u.Env.Exec.Run(ctx, "apt-get", "install", "-y", "unattended-upgrades"); err != nil {
			return nil, fmt.Errorf("apt install unattended-upgrades: %w (%s)", err, string(out))
		}
		_, _ = u.Env.Exec.Run(ctx, "systemctl", "enable", "--now", "unattended-upgrades")
	}
	b, err := json.Marshal(map[string]string{
		"conf_path": confPath,
		"previous":  prev,
		"stamp":     stamp,
	})
	if err != nil {
		return nil, err
	}
	return RevertPayload(b), nil
}

func (u *UnattendedEnable) Verify(_ context.Context) error {
	if _, err := os.Stat(u.Env.UnattendedStampPath()); err != nil {
		return fmt.Errorf("stamp missing: %w", err)
	}
	return nil
}

func (u *UnattendedEnable) Revert(_ context.Context, payload RevertPayload) error {
	var meta struct {
		ConfPath string `json:"conf_path"`
		Previous string `json:"previous"`
		Stamp    string `json:"stamp"`
	}
	if err := json.Unmarshal(payload, &meta); err != nil {
		return err
	}
	_ = os.Remove(meta.Stamp)
	if meta.Previous == "" {
		return os.Remove(meta.ConfPath)
	}
	return u.Env.WriteFile(meta.ConfPath, []byte(meta.Previous), 0o644)
}

func (u *UnattendedEnable) RequiresLockoutConfirm() bool { return false }
