package hostops

import (
	"context"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/proxybuild"
)

func TestProxySwapPlan(t *testing.T) {
	proxybuild.SetPanelVersion("0.7.21")
	cfg := config.Load()
	act := NewProxySwap(cfg)
	plan, err := act.Plan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if plan["target_image"] != "ghcr.io/joeydekruis/dockyrd-proxy:v0.7.21" {
		t.Fatalf("target: %v", plan["target_image"])
	}
	if act.RequiresLockoutConfirm() {
		t.Fatal("proxy.swap should not require lockout confirm")
	}
}