package hostops

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"
	"time"
)

const DefaultLockoutConfirmSeconds = 120

// LockoutTimer schedules automatic revert after a confirm deadline.
type LockoutTimer interface {
	Schedule(ctx context.Context, actionID string, notAfter time.Time) error
	Cancel(ctx context.Context, actionID string) error
}

// NoopLockoutTimer is used when lockout scheduling is disabled.
type NoopLockoutTimer struct{}

func (NoopLockoutTimer) Schedule(context.Context, string, time.Time) error { return nil }
func (NoopLockoutTimer) Cancel(context.Context, string) error            { return nil }

// MemoryLockoutTimer records schedules for unit tests.
type MemoryLockoutTimer struct {
	mu      sync.Mutex
	byID    map[string]time.Time
	cancelled map[string]bool
}

func NewMemoryLockoutTimer() *MemoryLockoutTimer {
	return &MemoryLockoutTimer{
		byID:      map[string]time.Time{},
		cancelled: map[string]bool{},
	}
}

func (m *MemoryLockoutTimer) Schedule(_ context.Context, actionID string, notAfter time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byID[actionID] = notAfter
	delete(m.cancelled, actionID)
	return nil
}

func (m *MemoryLockoutTimer) Cancel(_ context.Context, actionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancelled[actionID] = true
	return nil
}

func (m *MemoryLockoutTimer) ScheduledAt(actionID string) (time.Time, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.byID[actionID]
	return t, ok
}

func (m *MemoryLockoutTimer) WasCancelled(actionID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cancelled[actionID]
}

// SystemdLockoutTimer schedules dockyard host revert via systemd-run.
//
// Requires dockyard on PATH as installed by packaging; timer runs as root (same as panel today).
// Uses systemd-run --on-active=<duration> --unit=dockyard-host-revert-<id> dockyard host revert <id>.
type SystemdLockoutTimer struct {
	DockyardBin string
	Run         func(ctx context.Context, name string, args ...string) error
}

func NewSystemdLockoutTimer() *SystemdLockoutTimer {
	bin := "dockyard"
	return &SystemdLockoutTimer{
		DockyardBin: bin,
		Run:         execSystemdRun,
	}
}

func execSystemdRun(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Debug("systemd-run failed", "err", err, "output", string(out))
		return fmt.Errorf("systemd-run: %w", err)
	}
	return nil
}

func (s *SystemdLockoutTimer) Schedule(ctx context.Context, actionID string, notAfter time.Time) error {
	delay := time.Until(notAfter)
	if delay < time.Second {
		delay = time.Second
	}
	secs := int64(delay.Round(time.Second) / time.Second)
	unit := fmt.Sprintf("dockyard-host-revert-%s", actionID)
	args := []string{
		"--on-active=" + fmt.Sprintf("%d", secs),
		"--unit=" + unit,
		"--collect",
		s.DockyardBin, "host", "revert", actionID,
	}
	return s.Run(ctx, "systemd-run", args...)
}

func (s *SystemdLockoutTimer) Cancel(ctx context.Context, actionID string) error {
	unit := fmt.Sprintf("dockyard-host-revert-%s", actionID)
	_ = s.Run(ctx, "systemctl", "stop", unit)
	_ = s.Run(ctx, "systemctl", "reset-failed", unit)
	return nil
}
