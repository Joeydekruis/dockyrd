package config

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	legacyDataDir       = "/var/lib/dockyard"
	legacyConfigDir     = "/etc/dockyard"
	legacyEnvDataDir    = "DOCKYARD_DATA_DIR"
	legacyEnvConfigDir  = "DOCKYARD_CONFIG_DIR"
	legacyEnvListen     = "DOCKYARD_LISTEN"
	legacyEnvPublicIPv4 = "DOCKYARD_PUBLIC_IPV4"
	legacyEnvPublicURL  = "DOCKYARD_PUBLIC_URL"
)

func applyLegacyEnvAliases(cfg *Config) {
	if os.Getenv(envDataDir) == "" {
		if v := os.Getenv(legacyEnvDataDir); v != "" {
			cfg.DataDir = v
		}
	}
	if os.Getenv(envConfigDir) == "" {
		if v := os.Getenv(legacyEnvConfigDir); v != "" {
			cfg.ConfigDir = v
		}
	}
	if os.Getenv(envListen) == "" {
		if v := os.Getenv(legacyEnvListen); v != "" {
			cfg.Listen = v
		}
	}
	if os.Getenv(envPublicIPv4) == "" {
		if v := os.Getenv(legacyEnvPublicIPv4); v != "" {
			cfg.PublicIPv4 = v
		}
	}
}

// resolveLegacyInstallPaths points at pre-rebrand dirs when a fresh empty DB was
// created under /var/lib/dockyrd after self-update while the real data still lives
// under /var/lib/dockyard.
func resolveLegacyInstallPaths(cfg Config) Config {
	if os.Getenv(envDataDir) != "" || os.Getenv(legacyEnvDataDir) != "" {
		return cfg
	}
	newDB := cfg.DatabasePath()
	legacyDB := filepath.Join(legacyDataDir, "control.db")
	newOK := controlDBUsable(newDB)
	legacyOK := controlDBUsable(legacyDB)
	if !legacyOK {
		return cfg
	}
	if !newOK {
		cfg.DataDir = legacyDataDir
		cfg.ConfigDir = resolveLegacyConfigDir(cfg.ConfigDir)
		return cfg
	}
	newUsers := countUsers(newDB)
	legacyUsers := countUsers(legacyDB)
	if newUsers == 0 && legacyUsers > 0 {
		cfg.DataDir = legacyDataDir
		cfg.ConfigDir = resolveLegacyConfigDir(cfg.ConfigDir)
	}
	return cfg
}

func resolveLegacyConfigDir(current string) string {
	if os.Getenv(envConfigDir) != "" || os.Getenv(legacyEnvConfigDir) != "" {
		return current
	}
	key := filepath.Join(legacyConfigDir, "secret.key")
	if _, err := os.Stat(key); err == nil {
		return legacyConfigDir
	}
	return current
}

func controlDBUsable(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return countUsers(path) >= 0
}

func countUsers(dbPath string) int {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(1000)")
	if err != nil {
		return -1
	}
	defer db.Close()
	var n int
	err = db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	if err != nil {
		return -1
	}
	return n
}
