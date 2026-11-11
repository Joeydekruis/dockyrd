package securitychecks

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/hostops"
)

func TestCheckProxyDefaultsPasses(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
	r := checkProxyDefaults(Deps{Cfg: cfg, Env: hostops.NewHostEnv(cfg)}, time.Now().Unix())
	if r.Status != StatusOK {
		t.Fatalf("status=%s detail=%s", r.Status, r.Detail)
	}
}
