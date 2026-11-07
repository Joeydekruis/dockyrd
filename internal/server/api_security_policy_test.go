package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestSecurityPolicyAdminOnly(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	mux := http.NewServeMux()
	registerAPIRoutes(mux, apiDeps{cfg: cfg, q: q, db: conn})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/security/policy", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauth=%d", rec.Code)
	}
}

func TestSecurityPolicyGetAfterSetup(t *testing.T) {
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
	mux := http.NewServeMux()
	registerAPIRoutes(mux, apiDeps{cfg: cfg, q: q, db: conn})
	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Owner", "email": "o@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/security/policy", nil)
	for _, c := range recSetup.Result().Cookies() {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get policy=%d %s", rec.Code, rec.Body.String())
	}
}

func TestServiceSecurityEventsEmpty(t *testing.T) {
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
	mux := http.NewServeMux()
	registerAPIRoutes(mux, apiDeps{cfg: cfg, q: q, db: conn})
	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Owner", "email": "o@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/services/does-not-exist/security/events", nil)
	for _, c := range recSetup.Result().Cookies() {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("events=%d", rec.Code)
	}
}
