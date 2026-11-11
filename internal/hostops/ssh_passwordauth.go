package hostops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

)

const sshDropInBody = `# Managed by dockyrd — do not edit
PasswordAuthentication no
KbdInteractiveAuthentication no
PermitRootLogin prohibit-password
`

// SSHPasswordAuthDisable writes sshd drop-in and reloads sshd.
type SSHPasswordAuthDisable struct {
	Env        HostEnv
	SkipPrechk bool // for tests only
}

func NewSSHPasswordAuthDisable(env HostEnv) *SSHPasswordAuthDisable {
	return &SSHPasswordAuthDisable{Env: env}
}

func (s *SSHPasswordAuthDisable) Kind() string { return KindSSHPasswordAuthDisable }

func (s *SSHPasswordAuthDisable) Plan(ctx context.Context) (Plan, error) {
	if !s.SkipPrechk {
		if err := s.precheck(ctx); err != nil {
			return nil, err
		}
	}
	return Plan{
		"summary":  "Disable SSH password authentication via drop-in",
		"path":     s.Env.SSHDropInPath(),
		"diff":     unifiedDiff("(none)", sshDropInBody),
		"drop_in":  sshDropInBody,
		"lockout":  "Confirm within 120s or changes revert automatically",
	}, nil
}

func (s *SSHPasswordAuthDisable) precheck(ctx context.Context) error {
	if s.Env.HostRoot() != "" {
		return nil
	}
	return sshPrecheck(ctx, s.Env.Exec, s.SkipPrechk)
}

func (s *SSHPasswordAuthDisable) Apply(ctx context.Context) (RevertPayload, error) {
	path := s.Env.SSHDropInPath()
	prev := ""
	if data, err := s.Env.ReadFile(path); err == nil {
		prev = string(data)
	}
	if err := s.Env.WriteFile(path, []byte(sshDropInBody), 0o644); err != nil {
		return nil, err
	}
	if s.Env.HostRoot() == "" {
		if out, err := s.Env.Exec.Run(ctx, "sshd", "-t"); err != nil {
			_ = restoreHostFile(s.Env, path, prev)
			return nil, fmt.Errorf("sshd -t: %w (%s)", err, strings.TrimSpace(string(out)))
		}
		if out, err := s.Env.Exec.Run(ctx, "systemctl", "reload", "sshd"); err != nil {
			if out2, err2 := s.Env.Exec.Run(ctx, "systemctl", "reload", "ssh"); err2 != nil {
				_ = restoreHostFile(s.Env, path, prev)
				return nil, fmt.Errorf("reload sshd: %w (%s)", err, string(out)+string(out2))
			}
		}
	}
	b, err := json.Marshal(map[string]string{"path": path, "previous": prev})
	if err != nil {
		return nil, err
	}
	return RevertPayload(b), nil
}

func (s *SSHPasswordAuthDisable) Verify(ctx context.Context) error {
	if _, err := s.Env.ReadFile(s.Env.SSHDropInPath()); err != nil {
		return err
	}
	if s.Env.HostRoot() != "" {
		return nil
	}
	enabled, err := SSHPasswordAuthEnabled(ctx, s.Env.Exec)
	if err != nil {
		return err
	}
	if enabled {
		return fmt.Errorf("password authentication still enabled")
	}
	return nil
}

func (s *SSHPasswordAuthDisable) Revert(_ context.Context, payload RevertPayload) error {
	var meta struct {
		Path     string `json:"path"`
		Previous string `json:"previous"`
	}
	if err := json.Unmarshal(payload, &meta); err != nil {
		return err
	}
	if meta.Previous == "" {
		return os.Remove(meta.Path)
	}
	return s.Env.WriteFile(meta.Path, []byte(meta.Previous), 0o644)
}

func (s *SSHPasswordAuthDisable) RequiresLockoutConfirm() bool { return true }

func unifiedDiff(before, after string) string {
	return fmt.Sprintf("--- before\n+++ after\n%s", after)
}
