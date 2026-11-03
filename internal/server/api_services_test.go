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

func TestCreatePostgresService(t *testing.T) {
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

	mux := http.NewServeMux()
	deps := apiDeps{cfg: cfg, q: q}
	registerAPIRoutes(mux, deps)

	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Joey", "email": "o@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	reqSetup := httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody))
	mux.ServeHTTP(recSetup, reqSetup)
	if recSetup.Code != http.StatusOK {
		t.Fatalf("setup: %d %s", recSetup.Code, recSetup.Body.String())
	}
	cookies := recSetup.Result().Cookies()

	projBody, _ := json.Marshal(map[string]string{"name": "App"})
	recProj := httptest.NewRecorder()
	reqProj := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(projBody))
	for _, c := range cookies {
		reqProj.AddCookie(c)
	}
	mux.ServeHTTP(recProj, reqProj)
	var proj projectResponse
	_ = json.NewDecoder(recProj.Body).Decode(&proj)

	svcBody, _ := json.Marshal(map[string]any{
		"project_id": proj.ID,
		"kind":       "postgres",
		"name":       "Primary DB",
		"config":     map[string]any{"version": 17},
	})
	recSvc := httptest.NewRecorder()
	reqSvc := httptest.NewRequest(http.MethodPost, "/api/v1/services", bytes.NewReader(svcBody))
	for _, c := range cookies {
		reqSvc.AddCookie(c)
	}
	mux.ServeHTTP(recSvc, reqSvc)
	if recSvc.Code != http.StatusCreated {
		t.Fatalf("create service: %d %s", recSvc.Code, recSvc.Body.String())
	}
	var svc serviceResponse
	if err := json.NewDecoder(recSvc.Body).Decode(&svc); err != nil {
		t.Fatal(err)
	}
	if svc.Kind != "postgres" || svc.Slug == "" {
		t.Fatalf("unexpected service: %+v", svc)
	}
	var cfgCheck struct {
		PasswordEnc string `json:"password_enc"`
		Version     int    `json:"version"`
	}
	if err := json.Unmarshal(svc.Config, &cfgCheck); err != nil {
		t.Fatal(err)
	}
	if cfgCheck.PasswordEnc != "" {
		t.Fatal("expected redacted config in API response")
	}
	if cfgCheck.Version != 17 {
		t.Fatalf("version: %d", cfgCheck.Version)
	}

	recConn := httptest.NewRecorder()
	reqConn := httptest.NewRequest(http.MethodGet, "/api/v1/services/"+svc.ID+"/connection?reveal=1", nil)
	for _, c := range cookies {
		reqConn.AddCookie(c)
	}
	mux.ServeHTTP(recConn, reqConn)
	if recConn.Code != http.StatusOK {
		t.Fatalf("connection: %d %s", recConn.Code, recConn.Body.String())
	}
	var connResp struct {
		Variables map[string]string `json:"variables"`
	}
	if err := json.NewDecoder(recConn.Body).Decode(&connResp); err != nil {
		t.Fatal(err)
	}
	if connResp.Variables["DATABASE_URL"] == "" || connResp.Variables["PGHOST"] != svc.Slug {
		t.Fatalf("vars: %+v", connResp.Variables)
	}
}
