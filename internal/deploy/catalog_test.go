package deploy

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/refs"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestResolveVariableAgainstProjectCatalog(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:   filepath.Join(dir, "data"),
		ConfigDir: filepath.Join(dir, "etc"),
	}
	if _, err := setup.Init(cfg); err != nil {
		t.Fatal(err)
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	ctx := context.Background()
	srv, _ := q.GetLocalServer(ctx)
	now := int64(1)
	projID := "01TESTPROJECT000000000000"
	envID := "01TESTENV000000000000000"
	_ = q.CreateProject(ctx, sqlc.CreateProjectParams{
		ID: projID, ServerID: srv.ID, Name: "P", Slug: "p",
		CreatedAt: sql.NullInt64{Int64: now, Valid: true},
		UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
	})
	_ = q.CreateEnvironment(ctx, sqlc.CreateEnvironmentParams{ID: envID, ProjectID: projID, Name: "production"})
	key, _ := crypto.LoadMasterKey(cfg.SecretKeyPath())
	pgCfg, _ := dbkind.NewPostgresConfigJSON(key, 17)
	dbID := "01TESTDB0000000000000000"
	_ = q.CreateService(ctx, sqlc.CreateServiceParams{
		ID: dbID, EnvironmentID: envID, Kind: "postgres", Name: "DB", Slug: "db",
		DesiredState: "running", Config: pgCfg,
		CreatedAt: sql.NullInt64{Int64: now, Valid: true},
		UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
	})

	rt := &Runtime{Cfg: cfg, Q: q}
	catalog, err := rt.projectCatalog(ctx, projID)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := refs.Resolve("${{db.DATABASE_URL}}", refs.Catalog(catalog))
	if err != nil {
		t.Fatal(err)
	}
	if resolved == "" || resolved == "${{db.DATABASE_URL}}" {
		t.Fatalf("got %q", resolved)
	}
	pg, pw, _ := dbkind.ParsePostgres(pgCfg, key)
	want := dbkind.Published("postgres", "db", pw, pg)["DATABASE_URL"]
	if resolved != want {
		t.Fatalf("want %q got %q", want, resolved)
	}
}

func TestPublishedAppIncludesPrivateHost(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{"port": 3000})
	vars, err := publishedForService(context.Background(), nil, nil, sqlc.Service{
		Kind: "app", Slug: "web", ID: "x",
	}, raw)
	if err != nil {
		t.Fatal(err)
	}
	if vars["PRIVATE_HOST"] != "web:3000" {
		t.Fatalf("%+v", vars)
	}
}
