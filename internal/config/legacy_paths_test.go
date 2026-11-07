package config

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestLegacyEnvAliases(t *testing.T) {
	t.Setenv(envDataDir, "")
	t.Setenv(legacyEnvDataDir, "/var/lib/custom-dockyard")
	cfg := Config{DataDir: defaultDataDir}
	applyLegacyEnvAliases(&cfg)
	if cfg.DataDir != "/var/lib/custom-dockyard" {
		t.Fatalf("DataDir=%q", cfg.DataDir)
	}
}

func TestCountUsersEmpty(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "control.db")
	if err := writeMiniUsersDB(dbPath, 0); err != nil {
		t.Fatal(err)
	}
	if got := countUsers(dbPath); got != 0 {
		t.Fatalf("count=%d", got)
	}
}

func TestCountUsersPopulated(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "control.db")
	if err := writeMiniUsersDB(dbPath, 2); err != nil {
		t.Fatal(err)
	}
	if got := countUsers(dbPath); got != 2 {
		t.Fatalf("count=%d", got)
	}
}

func writeMiniUsersDB(path string, users int) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY)`); err != nil {
		return err
	}
	for i := 0; i < users; i++ {
		id := fmt.Sprintf("%s-u%d", filepath.Base(path), i)
		if _, err := db.Exec(`INSERT INTO users (id) VALUES (?)`, id); err != nil {
			return err
		}
	}
	return nil
}
