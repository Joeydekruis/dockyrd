package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func TestMigration00011ComposeKind(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "control.db")
	conn, q, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	ctx := context.Background()
	envID := "env-test"
	if err := q.CreateEnvironment(ctx, sqlc.CreateEnvironmentParams{
		ID:        envID,
		ProjectID: seedProject(t, ctx, q),
		Name:      "production",
	}); err != nil {
		t.Fatalf("CreateEnvironment: %v", err)
	}
	if err := q.CreateService(ctx, sqlc.CreateServiceParams{
		ID:            "svc-compose-test",
		EnvironmentID: envID,
		Kind:          "compose",
		Name:          "Stack",
		Slug:          "stack",
		DesiredState:  "running",
		Config:        []byte(`{"source":{"type":"inline","yaml":"services:\n  web:\n    image: nginx:alpine\n"},"project_name":"dyc-svccompo"}`),
	}); err != nil {
		t.Fatalf("CreateService compose: %v", err)
	}
}

func seedProject(t *testing.T, ctx context.Context, q *sqlc.Queries) string {
	t.Helper()
	serverID := "srv-local"
	_ = q.CreateServer(ctx, sqlc.CreateServerParams{
		ID:   serverID,
		Name: sql.NullString{String: "local", Valid: true},
	})
	pid := "proj-test"
	_ = q.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:       pid,
		ServerID: serverID,
		Name:     "Test",
		Slug:     "test",
	})
	return pid
}
