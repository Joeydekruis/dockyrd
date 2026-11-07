package hostops

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func testRunner(t *testing.T, cfg config.Config, reg *Registry, timer LockoutTimer) *Runner {
	t.Helper()
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if reg == nil {
		reg = NewRegistry()
		reg.Register(KindTestEcho, func() Action { return NewTestEcho(cfg) })
	}
	return NewRunner(conn, q, reg, audit.Logger{Q: q}, timer)
}

func TestTestEchoHappyPath(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, ConfigDir: filepath.Join(dir, "etc")}
	_, _ = setup.Init(cfg)

	r := testRunner(t, cfg, nil, NewMemoryLockoutTimer())
	ctx := context.Background()

	id, err := r.Plan(ctx, KindTestEcho, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Apply(ctx, id, ""); err != nil {
		t.Fatal(err)
	}
	if err := r.Verify(ctx, id); err != nil {
		t.Fatal(err)
	}
	row, err := r.Q.GetHostAction(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if row.State != string(StateConfirmed) {
		t.Fatalf("state=%s want confirmed", row.State)
	}
}

func TestVerifyFailureReverts(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, ConfigDir: filepath.Join(dir, "etc")}
	_, _ = setup.Init(cfg)

	reg := NewRegistry()
	reg.Register(KindTestEcho, func() Action {
		return &fakeVerifyFail{inner: NewTestEcho(cfg)}
	})
	r := testRunner(t, cfg, reg, NewMemoryLockoutTimer())
	ctx := context.Background()

	id, err := r.Plan(ctx, KindTestEcho, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Apply(ctx, id, ""); err != nil {
		t.Fatal(err)
	}
	_ = r.Verify(ctx, id)
	row, _ := r.Q.GetHostAction(ctx, id)
	if row.State != string(StateReverted) {
		t.Fatalf("state=%s want reverted", row.State)
	}
}

type fakeVerifyFail struct {
	inner *TestEcho
}

func (f *fakeVerifyFail) Kind() string                       { return f.inner.Kind() }
func (f *fakeVerifyFail) Plan(ctx context.Context) (Plan, error) { return f.inner.Plan(ctx) }
func (f *fakeVerifyFail) Apply(ctx context.Context) (RevertPayload, error) {
	return f.inner.Apply(ctx)
}
func (f *fakeVerifyFail) Verify(context.Context) error { return errors.New("verify boom") }
func (f *fakeVerifyFail) Revert(ctx context.Context, p RevertPayload) error {
	return f.inner.Revert(ctx, p)
}
func (f *fakeVerifyFail) RequiresLockoutConfirm() bool { return false }

func TestLockoutConfirmPath(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, ConfigDir: filepath.Join(dir, "etc")}
	_, _ = setup.Init(cfg)

	reg := NewRegistry()
	reg.Register(KindTestEcho, func() Action { return NewTestEcho(cfg).WithLockoutConfirm() })
	timer := NewMemoryLockoutTimer()
	r := testRunner(t, cfg, reg, timer)
	ctx := context.Background()

	id, _ := r.Plan(ctx, KindTestEcho, "")
	_ = r.Apply(ctx, id, "")
	_ = r.Verify(ctx, id)

	row, _ := r.Q.GetHostAction(ctx, id)
	if row.State != string(StateAwaitingConfirm) {
		t.Fatalf("state=%s", row.State)
	}
	if _, ok := timer.ScheduledAt(id); !ok {
		t.Fatal("timer not scheduled")
	}
	if err := r.Confirm(ctx, id, ""); err != nil {
		t.Fatal(err)
	}
	if !timer.WasCancelled(id) {
		t.Fatal("timer not cancelled on confirm")
	}
	row, _ = r.Q.GetHostAction(ctx, id)
	if row.State != string(StateConfirmed) {
		t.Fatalf("state=%s", row.State)
	}
}

func TestRevertIdempotent(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, ConfigDir: filepath.Join(dir, "etc")}
	_, _ = setup.Init(cfg)
	r := testRunner(t, cfg, nil, NewMemoryLockoutTimer())
	ctx := context.Background()
	id, _ := r.Plan(ctx, KindTestEcho, "")
	_ = r.Apply(ctx, id, "")
	if err := r.Revert(ctx, id, RevertMeta{Source: "test"}); err != nil {
		t.Fatal(err)
	}
	if err := r.Revert(ctx, id, RevertMeta{Source: "test"}); err != nil {
		t.Fatal("second revert should no-op", err)
	}
}

func TestStubApplyNotImplemented(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, ConfigDir: filepath.Join(dir, "etc")}
	_, _ = setup.Init(cfg)
	reg := DefaultRegistry(cfg)
	r := testRunner(t, cfg, reg, nil)
	ctx := context.Background()
	id, err := r.Plan(ctx, KindCrowdsecBouncerInstall, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Apply(ctx, id, ""); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("apply err=%v want ErrNotImplemented", err)
	}
}

func TestReconcileExpired(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, ConfigDir: filepath.Join(dir, "etc")}
	_, _ = setup.Init(cfg)
	reg := NewRegistry()
	reg.Register(KindTestEcho, func() Action { return NewTestEcho(cfg).WithLockoutConfirm() })
	r := testRunner(t, cfg, reg, NewMemoryLockoutTimer())
	r.Now = func() time.Time { return time.Unix(1_700_000_000, 0) }
	r.Lockout = 60 * time.Second

	ctx := context.Background()
	id, _ := r.Plan(ctx, KindTestEcho, "")
	_ = r.Apply(ctx, id, "")
	_ = r.Verify(ctx, id)

	r.Now = func() time.Time { return time.Unix(1_700_000_200, 0) }
	n, err := Reconcile(ctx, r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("reverted=%d want 1", n)
	}
	row, _ := r.Q.GetHostAction(ctx, id)
	if row.State != string(StateReverted) {
		t.Fatalf("state=%s", row.State)
	}
}

func TestMarkerFileLifecycle(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, ConfigDir: filepath.Join(dir, "etc")}
	_, _ = setup.Init(cfg)
	r := testRunner(t, cfg, nil, nil)
	ctx := context.Background()
	id, _ := r.Plan(ctx, KindTestEcho, "")
	_ = r.Apply(ctx, id, "")
	path := filepath.Join(dir, "hostops-test-echo.marker")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if err := r.Revert(ctx, id, RevertMeta{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("marker should be removed")
	}
}
