package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultListen    = ":8000"
	defaultDataDir   = "/var/lib/dockyard"
	defaultConfigDir = "/etc/dockyard"
	envDataDir       = "DOCKYARD_DATA_DIR"
	envConfigDir     = "DOCKYARD_CONFIG_DIR"
	envListen        = "DOCKYARD_LISTEN"
	envPublicIPv4    = "DOCKYARD_PUBLIC_IPV4"
	envPublicURL     = "DOCKYARD_PUBLIC_URL"
)

// Config holds process-level paths and network binding.
type Config struct {
	DataDir     string
	ConfigDir   string
	Listen      string
	PublicIPv4  string
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
	if v := os.Getenv(envPublicIPv4); v != "" {
		cfg.PublicIPv4 = v
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

// CaddyAdminSocket is the host path to dy-proxy admin unix socket.
func (c Config) CaddyAdminSocket() string {
	return filepath.Join(c.DataDir, "caddy", "admin.sock")
}

// BuildsDir is the root for ephemeral clone/build workspaces.
func (c Config) BuildsDir() string {
	return filepath.Join(c.DataDir, "builds")
}

// SitesDir holds static site releases and current symlinks.
func (c Config) SitesDir() string {
	return filepath.Join(c.DataDir, "sites")
}

// EnvFilePath is the optional systemd EnvironmentFile.
func (c Config) EnvFilePath() string {
	return filepath.Join(c.ConfigDir, "dockyard.env")
}

const (
	defaultBuildTimeout = 20 * time.Minute
	envBuildTimeout     = "DOCKYARD_BUILD_TIMEOUT"
	envMinFreeDiskPct   = "DOCKYARD_MIN_FREE_DISK_PERCENT"
)

// BuildTimeout is the maximum duration for clone/build steps (default 20m).
func (c Config) BuildTimeout() time.Duration {
	if v := strings.TrimSpace(os.Getenv(envBuildTimeout)); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return defaultBuildTimeout
}

// PublicURL is the external HTTPS origin for links in email (optional env override).
func (c Config) PublicURL() string {
	v := strings.TrimSpace(os.Getenv(envPublicURL))
	for len(v) > 0 && v[len(v)-1] == '/' {
		v = v[:len(v)-1]
	}
	return v
}

// MinFreeDiskPercent is the minimum free disk ratio required before builds (default 10).
func (c Config) MinFreeDiskPercent() int {
	if v := strings.TrimSpace(os.Getenv(envMinFreeDiskPct)); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			return n
		}
	}
	return 10
}
