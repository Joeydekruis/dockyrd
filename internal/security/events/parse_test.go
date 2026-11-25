package events

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

const fixtureLine = `{"transaction":{"client_ip":"203.0.113.10","timestamp":"2026-11-25T12:00:00.000Z","request":{"uri":"/api/foo?x=1","headers":{"Host":["app.example.com"]}}},"messages":[{"message":"SQL Injection Attack Detected via libinjection","data":{"id":942100,"severity":"2"}}]}`

func TestParseCorazaAuditLine(t *testing.T) {
	hit, err := ParseCorazaAuditLine(fixtureLine)
	if err != nil {
		t.Fatal(err)
	}
	if hit.Hostname != "app.example.com" {
		t.Fatalf("host=%q", hit.Hostname)
	}
	if hit.Path != "/api/foo?x=1" {
		t.Fatalf("path=%q", hit.Path)
	}
	if hit.IP != "203.0.113.10" {
		t.Fatalf("ip=%q", hit.IP)
	}
	if hit.Rule != "942100" {
		t.Fatalf("rule=%q", hit.Rule)
	}
	if hit.Action != "detect" {
		t.Fatalf("action=%q", hit.Action)
	}
}

func TestInsertAndPurge(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
	if _, err := setup.Init(cfg); err != nil {
		t.Fatal(err)
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	ctx := t.Context()
	now := time.Now().UTC()
	ts := sql.NullInt64{Int64: now.Unix(), Valid: true}
	srvID := "01TESTSERVWAF000000000001"
	projID := "01TESTPROJWAF00000000001"
	envID := "01TESTENVWAF000000000001"
	svcID := "01TESTSVCWAF000000000001"

	if err := q.CreateServer(ctx, sqlc.CreateServerParams{ID: srvID, CreatedAt: ts}); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateProject(ctx, sqlc.CreateProjectParams{
		ID: projID, ServerID: srvID, Name: "P", Slug: "p-waf", CreatedAt: ts, UpdatedAt: ts,
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateEnvironment(ctx, sqlc.CreateEnvironmentParams{ID: envID, ProjectID: projID, Name: "production"}); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateService(ctx, sqlc.CreateServiceParams{
		ID: svcID, EnvironmentID: envID, Kind: "app", Name: "web", Slug: "web",
		DesiredState: "running", Config: "{}", CreatedAt: ts, UpdatedAt: ts,
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateDomain(ctx, sqlc.CreateDomainParams{
		ID: "dom_waf", ServiceID: svcID, Hostname: "app.example.com",
	}); err != nil {
		t.Fatal(err)
	}

	hit, err := ParseCorazaAuditLine(fixtureLine)
	if err != nil {
		t.Fatal(err)
	}
	ins := Inserter{Q: q}
	row, err := ins.InsertWAFHit(ctx, hit)
	if err != nil {
		t.Fatal(err)
	}
	if row.Source != "waf" || row.ServiceID != svcID {
		t.Fatalf("%+v", row)
	}

	old := hit
	old.TS = now.Add(-10 * 24 * time.Hour)
	old.Path = "/old"
	if _, err := ins.InsertWAFHit(ctx, old); err != nil {
		t.Fatal(err)
	}
	n, err := PurgeDefaultRetention(ctx, q, now)
	if err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatalf("expected purge of old row, got %d", n)
	}
	list, err := q.ListSecurityEventsByService(ctx, sqlc.ListSecurityEventsByServiceParams{
		ServiceID: svcID, Limit: 50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 remaining, got %d", len(list))
	}
}

func TestWorkerSkipsMissingLog(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
	if _, err := setup.Init(cfg); err != nil {
		t.Fatal(err)
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	w := &Worker{Cfg: cfg, Q: q, Now: func() time.Time { return time.Now().UTC() }}
	if err := w.TickOnce(t.Context()); err != nil {
		t.Fatal(err)
	}
}
