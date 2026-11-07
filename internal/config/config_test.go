package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadDefaultsListen(t *testing.T) {
	t.Setenv(envDataDir, "")
	t.Setenv(envListen, "")
	cfg := Load()
	if cfg.Listen != defaultListen {
		t.Fatalf("Listen = %q", cfg.Listen)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv(envDataDir, "/tmp/dy-test-data")
	t.Setenv(envConfigDir, "/tmp/dy-test-etc")
	t.Setenv(envListen, ":9090")
	cfg := Load()
	if cfg.DataDir != "/tmp/dy-test-data" {
		t.Fatalf("DataDir = %q", cfg.DataDir)
	}
	if cfg.ConfigDir != "/tmp/dy-test-etc" {
		t.Fatalf("ConfigDir = %q", cfg.ConfigDir)
	}
	if cfg.Listen != ":9090" {
		t.Fatalf("Listen = %q", cfg.Listen)
	}
	want := filepath.Join("/tmp/dy-test-data", "control.db")
	if cfg.DatabasePath() != want {
		t.Fatalf("DatabasePath = %q, want %q", cfg.DatabasePath(), want)
	}
	if cfg.SecretKeyPath() != "/tmp/dy-test-etc/secret.key" {
		t.Fatalf("SecretKeyPath = %q", cfg.SecretKeyPath())
	}
}

func TestDevDataDirNonLinux(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("linux uses /var/lib/dockyrd by default")
	}
	t.Setenv(envDataDir, "")
	t.Setenv(envConfigDir, "")
	cfg := Load()
	if cfg.DataDir == defaultDataDir {
		t.Fatalf("expected non-linux dev data dir, got %q", cfg.DataDir)
	}
	if cfg.ConfigDir == defaultConfigDir {
		t.Fatalf("expected non-linux dev config dir, got %q", cfg.ConfigDir)
	}
}
