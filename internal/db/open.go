package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// Open opens SQLite at path, runs migrations, and returns the pool plus queries.
func Open(path string) (*sql.DB, *sqlc.Queries, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, nil, fmt.Errorf("mkdir data dir: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1)

	if err := migrateUp(conn); err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	return conn, sqlc.New(conn), nil
}
