package securitychecks

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/hostops"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestRunAllIncludesEightChecks(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
	if _, err := setup.Init(cfg); err != nil {
		t.Fatal(err)
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	rep := RunAll(context.Background(), Deps{
		Cfg: cfg, Q: q, Env: hostops.NewHostEnv(cfg), Now: func() time.Time { return time.Unix(1_700_000_000, 0) },
	})
	if len(rep.Checks) != 8 {
		t.Fatalf("checks=%d want 8", len(rep.Checks))
	}
	seen := map[string]bool{}
	for _, c := range rep.Checks {
		seen[c.ID] = true
		if c.LastRunAt == 0 {
			t.Fatalf("missing last_run_at for %s", c.ID)
		}
	}
	for _, id := range []string{
		"firewall.active", "docker.published_ports", "ssh.password_auth", "updates.unattended",
		"reboot.pending", "panel.2fa", "panel.https", "proxy.defaults",
	} {
		if !seen[id] {
			t.Fatalf("missing %s", id)
		}
	}
}

func TestFixKindForCheck(t *testing.T) {
	if k, ok := FixKindForCheck("ssh.password_auth"); !ok || k != hostops.KindSSHPasswordAuthDisable {
		t.Fatalf("kind=%s ok=%v", k, ok)
	}
}
