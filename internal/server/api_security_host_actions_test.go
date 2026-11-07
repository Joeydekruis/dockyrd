package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/hostops"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestHostActionConfirmRequiresAdmin(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:   filepath.Join(dir, "data"),
		ConfigDir: filepath.Join(dir, "etc"),
	}
	initRes, err := setup.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	runner := hostops.NewRunner(conn, q, hostops.DefaultRegistry(cfg), audit.Logger{Q: q}, hostops.NewMemoryLockoutTimer())
	ctx := context.Background()
	reg := hostops.NewRegistry()
	reg.Register(hostops.KindTestEcho, func() hostops.Action {
		return hostops.NewTestEcho(cfg).WithLockoutConfirm()
	})
	runner.Registry = reg
	id, _ := runner.Plan(ctx, hostops.KindTestEcho, "")
	_ = runner.Apply(ctx, id, "")
	_ = runner.Verify(ctx, id)

	mux := http.NewServeMux()
	deps := apiDeps{cfg: cfg, q: q, db: conn, audit: audit.Logger{Q: q}}
	registerAPIRoutes(mux, deps)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/security/host-actions/"+id+"/confirm", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauth=%d", rec.Code)
	}

	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Owner", "email": "o@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux2 := http.NewServeMux()
	registerAPIRoutes(mux2, deps)
	mux2.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))

	rec2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/security/host-actions/"+id+"/confirm", nil)
	for _, c := range recSetup.Result().Cookies() {
		req.AddCookie(c)
	}
	mux.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Fatalf("owner confirm=%d body=%s", rec2.Code, rec2.Body.String())
	}
}

func TestHostActionConfirmWrongState(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
	initRes, err := setup.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	runner := hostops.NewRunner(conn, q, hostops.DefaultRegistry(cfg), audit.Logger{Q: q}, nil)
	id, _ := runner.Plan(context.Background(), hostops.KindTestEcho, "")

	mux := http.NewServeMux()
	deps := apiDeps{cfg: cfg, q: q, db: conn, audit: audit.Logger{Q: q}}
	registerAPIRoutes(mux, deps)

	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Owner", "email": "o2@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/security/host-actions/"+id+"/confirm", nil)
	for _, c := range recSetup.Result().Cookies() {
		req.AddCookie(c)
	}
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status=%d want 409", rec.Code)
	}
}

func TestHostActionRevertNotFound(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
	initRes, _ := setup.Init(cfg)
	conn, q, _ := db.Open(cfg.DatabasePath())
	t.Cleanup(func() { _ = conn.Close() })

	mux := http.NewServeMux()
	deps := apiDeps{cfg: cfg, q: q, db: conn, audit: audit.Logger{Q: q}}
	registerAPIRoutes(mux, deps)

	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Owner", "email": "o3@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/security/host-actions/does-not-exist/revert", nil)
	for _, c := range recSetup.Result().Cookies() {
		req.AddCookie(c)
	}
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d", rec.Code)
	}
}
