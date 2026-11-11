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

const unattendedAutoUpgradesConf = `APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
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
		"summary":     "Enable unattended-upgrades (security only, no auto-reboot)",
		"distro":      distro,
		"apt_conf":    u.Env.UnattendedAptConfPath(),
		"auto_period": u.Env.UnattendedAutoUpgradesPath(),
		"stamp":       u.Env.UnattendedStampPath(),
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
	aptPath := u.Env.UnattendedAptConfPath()
	autoPath := u.Env.UnattendedAutoUpgradesPath()
	prevApt := readOptional(u.Env, aptPath)
	prevAuto := readOptional(u.Env, autoPath)
	if err := u.Env.WriteFile(aptPath, []byte(unattendedAptConf), 0o644); err != nil {
		return nil, err
	}
	if err := u.Env.WriteFile(autoPath, []byte(unattendedAutoUpgradesConf), 0o644); err != nil {
		_ = restoreHostFile(u.Env, aptPath, prevApt)
		return nil, err
	}
	stamp := u.Env.UnattendedStampPath()
	if err := u.Env.WriteFile(stamp, []byte("enabled\n"), 0o600); err != nil {
		_ = restoreHostFile(u.Env, aptPath, prevApt)
		_ = restoreHostFile(u.Env, autoPath, prevAuto)
		return nil, err
	}
	if u.Env.HostRoot() == "" {
		if out, err := u.Env.Exec.Run(ctx, "apt-get", "install", "-y", "unattended-upgrades"); err != nil {
			_ = u.revertFiles(stamp, aptPath, prevApt, autoPath, prevAuto)
			return nil, fmt.Errorf("apt install unattended-upgrades: %w (%s)", err, string(out))
		}
		_, _ = u.Env.Exec.Run(ctx, "systemctl", "enable", "--now", "unattended-upgrades")
	}
	b, err := json.Marshal(map[string]string{
		"apt_path":  aptPath,
		"auto_path": autoPath,
		"prev_apt":  prevApt,
		"prev_auto": prevAuto,
		"stamp":     stamp,
	})
	if err != nil {
		return nil, err
	}
	return RevertPayload(b), nil
}

func (u *UnattendedEnable) Verify(ctx context.Context) error {
	if _, err := os.Stat(u.Env.UnattendedStampPath()); err != nil {
		return fmt.Errorf("stamp missing: %w", err)
	}
	for _, p := range []string{u.Env.UnattendedAptConfPath(), u.Env.UnattendedAutoUpgradesPath()} {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("apt config missing %s: %w", filepath.Base(p), err)
		}
	}
	if u.Env.HostRoot() == "" {
		out, err := u.Env.Exec.Run(ctx, "systemctl", "is-active", "unattended-upgrades")
		if err != nil || strings.TrimSpace(string(out)) != "active" {
			return fmt.Errorf("unattended-upgrades service not active")
		}
	}
	return nil
}

func (u *UnattendedEnable) Revert(_ context.Context, payload RevertPayload) error {
	var meta struct {
		AptPath  string `json:"apt_path"`
		AutoPath string `json:"auto_path"`
		PrevApt  string `json:"prev_apt"`
		PrevAuto string `json:"prev_auto"`
		Stamp    string `json:"stamp"`
		// legacy payload keys
		ConfPath string `json:"conf_path"`
		Previous string `json:"previous"`
	}
	if err := json.Unmarshal(payload, &meta); err != nil {
		return err
	}
	if meta.AptPath == "" && meta.ConfPath != "" {
		meta.AptPath = meta.ConfPath
		meta.PrevApt = meta.Previous
	}
	_ = os.Remove(meta.Stamp)
	return u.revertFiles(meta.Stamp, meta.AptPath, meta.PrevApt, meta.AutoPath, meta.PrevAuto)
}

func (u *UnattendedEnable) revertFiles(stamp, aptPath, prevApt, autoPath, prevAuto string) error {
	_ = os.Remove(stamp)
	if err := restoreHostFile(u.Env, aptPath, prevApt); err != nil {
		return err
	}
	return restoreHostFile(u.Env, autoPath, prevAuto)
}

func (u *UnattendedEnable) RequiresLockoutConfirm() bool { return false }

func readOptional(env HostEnv, path string) string {
	data, err := env.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
