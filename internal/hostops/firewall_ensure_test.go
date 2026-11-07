package hostops

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

func TestFirewallEnsurePlanShape(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, ConfigDir: filepath.Join(dir, "etc")}
	env := NewHostEnv(cfg)
	a := NewFirewallEnsure(env)
	plan, err := a.Plan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if plan["summary"] == "" || plan["ruleset"] == "" {
		t.Fatalf("plan=%v", plan)
	}
}

func TestFirewallEnsureApplyVerifyDevHost(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, ConfigDir: filepath.Join(dir, "etc")}
	env := NewHostEnv(cfg)
	a := NewFirewallEnsure(env)
	ctx := context.Background()
	rev, err := a.Apply(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Verify(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.Revert(ctx, rev); err != nil {
		t.Fatal(err)
	}
}
