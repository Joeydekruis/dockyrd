package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	defaultListen   = ":8000"
	defaultDataDir  = "/var/lib/dockyard"
	envDataDir      = "DOCKYARD_DATA_DIR"
	envListen       = "DOCKYARD_LISTEN"
)

// Config holds process-level paths and network binding.
type Config struct {
	DataDir string
	Listen  string
}

// Load reads environment overrides with platform defaults.
func Load() Config {
	cfg := Config{
		DataDir: defaultDataDir,
		Listen:  defaultListen,
	}
	if v := os.Getenv(envDataDir); v != "" {
		cfg.DataDir = v
	}
	if v := os.Getenv(envListen); v != "" {
		cfg.Listen = v
	}
	if cfg.DataDir == defaultDataDir && runtime.GOOS != "linux" {
		cfg.DataDir = devDataDir()
	}
	return cfg
}

func devDataDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "dockyard")
	}
	return filepath.Join(os.TempDir(), "dockyard")
}

// DatabasePath returns the control-plane SQLite file path.
func (c Config) DatabasePath() string {
	return filepath.Join(c.DataDir, "control.db")
}
