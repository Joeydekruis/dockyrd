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
	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func instanceTestMux(t *testing.T) (*http.ServeMux, []*http.Cookie) {
	t.Helper()
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:    filepath.Join(dir, "data"),
		ConfigDir:  filepath.Join(dir, "etc"),
		PublicIPv4: "203.0.113.5",
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
	mux := http.NewServeMux()
	deps := apiDeps{cfg: cfg, q: q}
	registerAPIRoutes(mux, deps)

	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Joey", "email": "o@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))
	return mux, recSetup.Result().Cookies()
}

func TestVerifyWildcardRequiresPattern(t *testing.T) {
	mux, cookies := instanceTestMux(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/instance/verify-wildcard", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestInstanceSettingsIncludesSuggestedAppsDomain(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:    filepath.Join(dir, "data"),
		ConfigDir:  filepath.Join(dir, "etc"),
		PublicIPv4: "203.0.113.5",
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
	mux := http.NewServeMux()
	deps := apiDeps{cfg: cfg, q: q}
	registerAPIRoutes(mux, deps)

	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Joey", "email": "o@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))
	cookies := recSetup.Result().Cookies()

	if err := instance.SavePanel(t.Context(), q, instance.Panel{
		PublicURL: "https://panel.clientx.nl",
		Wildcard:  "*.apps.clientx.nl",
	}); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/instance", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	if payload["suggested_apps_domain"] != "*.apps.clientx.nl" {
		t.Fatalf("suggested=%v", payload["suggested_apps_domain"])
	}
}

func TestPatchWildcardResetsVerified(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:    filepath.Join(dir, "data"),
		ConfigDir:  filepath.Join(dir, "etc"),
		PublicIPv4: "203.0.113.5",
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
	mux := http.NewServeMux()
	deps := apiDeps{cfg: cfg, q: q}
	registerAPIRoutes(mux, deps)

	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Joey", "email": "o@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))
	cookies := recSetup.Result().Cookies()

	panel := instance.Panel{
		Wildcard:           "*.apps.a.example",
		WildcardVerifiedAt: 99,
	}
	instance.EnsureProbeHost(&panel)
	if err := instance.SavePanel(t.Context(), q, panel); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]string{"wildcard_domain": "*.apps.b.example"})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/settings/instance", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch=%d %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&out)
	verifiedAt, _ := out["wildcard_verified_at"].(float64)
	if verifiedAt != 0 {
		t.Fatalf("expected verified reset, got %v", out["wildcard_verified_at"])
	}
}
