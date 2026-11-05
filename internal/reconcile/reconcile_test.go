package reconcile

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func TestMarkInterruptedOnlyOnStartupTick(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:   filepath.Join(dir, "data"),
		ConfigDir: filepath.Join(dir, "etc"),
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	ctx := context.Background()
	now := time.Now().Unix()
	ts := sql.NullInt64{Int64: now, Valid: true}

	srvID := "01TESTSERVHOST000000000001"
	projID := "01TESTPROJ0000000000000001"
	envID := "01TESTENV00000000000000001"
	svcID := "01TESTSVC00000000000000001"
	depID := "01TESTDEP00000000000000001"

	if err := q.CreateServer(ctx, sqlc.CreateServerParams{ID: srvID, CreatedAt: ts}); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateProject(ctx, sqlc.CreateProjectParams{
		ID: srvID + "p", ServerID: srvID, Name: "P", Slug: "p", CreatedAt: ts, UpdatedAt: ts,
	}); err != nil {
		t.Fatal(err)
	}
	projID = srvID + "p"
	if err := q.CreateEnvironment(ctx, sqlc.CreateEnvironmentParams{ID: envID, ProjectID: projID, Name: "production"}); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateService(ctx, sqlc.CreateServiceParams{
		ID: svcID, EnvironmentID: envID, Kind: "app", Name: "web", Slug: "web",
		DesiredState: "running", Config: "{}", CreatedAt: ts, UpdatedAt: ts,
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateDeployment(ctx, sqlc.CreateDeploymentParams{
		ID: depID, ServiceID: svcID, Status: "building", Trigger: "manual", QueuedAt: ts,
	}); err != nil {
		t.Fatal(err)
	}

	loop := &Loop{Q: q, interruptedMarked: true}
	loop.Tick(ctx)
	loop.Tick(ctx)

	row, err := q.GetDeploymentByID(ctx, depID)
	if err != nil {
		t.Fatal(err)
	}
	if row.Status != "building" {
		t.Fatalf("status after ticks with interruptedMarked=true: %q want building", row.Status)
	}

	loop2 := &Loop{Q: q}
	loop2.Tick(ctx)
	row, _ = q.GetDeploymentByID(ctx, depID)
	if row.Status != "failed" {
		t.Fatalf("first startup tick: %q want failed", row.Status)
	}

	dep2 := "01TESTDEP00000000000000002"
	if err := q.CreateDeployment(ctx, sqlc.CreateDeploymentParams{
		ID: dep2, ServiceID: svcID, Status: "building", Trigger: "manual", QueuedAt: ts,
	}); err != nil {
		t.Fatal(err)
	}
	loop2.Tick(ctx)
	loop2.Tick(ctx)
	row2, err := q.GetDeploymentByID(ctx, dep2)
	if err != nil {
		t.Fatal(err)
	}
	if row2.Status != "building" {
		t.Fatalf("second+ tick must not mark building failed: %q", row2.Status)
	}
}
