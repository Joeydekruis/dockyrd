package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func migrateUp(conn *sql.DB) error {
	goose.SetBaseFS(migrationFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.Up(conn, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
