package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	defaultListen    = ":8000"
	defaultDataDir   = "/var/lib/dockyard"
	defaultConfigDir = "/etc/dockyard"
	envDataDir       = "DOCKYARD_DATA_DIR"
	envConfigDir     = "DOCKYARD_CONFIG_DIR"
	envListen        = "DOCKYARD_LISTEN"
)

// Config holds process-level paths and network binding.
type Config struct {
	DataDir   string
	ConfigDir string
	Listen    string
}

// Load reads environment overrides with platform defaults.
func Load() Config {
	cfg := Config{
		DataDir:   defaultDataDir,
		ConfigDir: defaultConfigDir,
		Listen:    defaultListen,
	}
	if v := os.Getenv(envDataDir); v != "" {
		cfg.DataDir = v
	}
	if v := os.Getenv(envConfigDir); v != "" {
		cfg.ConfigDir = v
	}
	if v := os.Getenv(envListen); v != "" {
		cfg.Listen = v
	}
	if runtime.GOOS == "linux" {
		return cfg
	}
	base := devBaseDir()
	if os.Getenv(envDataDir) == "" {
		cfg.DataDir = filepath.Join(base, "data")
	}
	if os.Getenv(envConfigDir) == "" {
		cfg.ConfigDir = filepath.Join(base, "etc")
	}
	return cfg
}

func devBaseDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "dockyard")
	}
	return filepath.Join(os.TempDir(), "dockyard")
}

// EnsureLayout creates config and data directories with safe permissions.
func (c Config) EnsureLayout() error {
	if err := os.MkdirAll(c.ConfigDir, 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(c.DataDir, 0o750); err != nil {
		return err
	}
	return nil
}

// DatabasePath returns the control-plane SQLite file path.
func (c Config) DatabasePath() string {
	return filepath.Join(c.DataDir, "control.db")
}

// SecretKeyPath returns the AES master key path.
func (c Config) SecretKeyPath() string {
	return filepath.Join(c.ConfigDir, "secret.key")
}

// EnvFilePath is the optional systemd EnvironmentFile.
func (c Config) EnvFilePath() string {
	return filepath.Join(c.ConfigDir, "dockyard.env")
}
