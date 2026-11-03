package doctor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

func TestReportHasFailures(t *testing.T) {
	r := Report{Checks: []Check{{Status: StatusWarn}}}
	if r.HasFailures() {
		t.Fatal("warn should not fail")
	}
	r.Checks = append(r.Checks, Check{Status: StatusFail})
	if !r.HasFailures() {
		t.Fatal("expected failure")
	}
}

func TestCheckLayoutMissingDataDir(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:   filepath.Join(dir, "missing-data"),
		ConfigDir: filepath.Join(dir, "etc"),
	}
	if err := os.MkdirAll(cfg.ConfigDir, 0o700); err != nil {
		t.Fatal(err)
	}
	rep := Run(context.Background(), cfg, "test")
	if !rep.HasFailures() {
		t.Fatal("expected failure for missing data dir")
	}
	var buf strings.Builder
	Print(&buf, rep)
	if !strings.Contains(buf.String(), "Data directory") {
		t.Fatalf("output: %s", buf.String())
	}
}
