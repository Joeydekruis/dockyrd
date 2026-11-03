package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// VacuumIntoBackup copies the SQLite database to dest using VACUUM INTO.
func VacuumIntoBackup(dbPath, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o750); err != nil {
		return err
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	defer conn.Close()
	query := fmt.Sprintf("VACUUM INTO %q", destPath)
	if _, err := conn.Exec(query); err != nil {
		return fmt.Errorf("vacuum into: %w", err)
	}
	return nil
}

// PreUpdateBackupPath returns a timestamped path under dataDir/backups.
func PreUpdateBackupPath(dataDir string, now time.Time) string {
	name := fmt.Sprintf("pre-update-%s.db", now.UTC().Format("20060102-150405"))
	return filepath.Join(dataDir, "backups", name)
}
