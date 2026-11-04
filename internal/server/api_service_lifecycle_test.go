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

func TestDeletePostgresService(t *testing.T) {
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

	delBody, _ := json.Marshal(map[string]any{
		"confirm_slug":  svc.Slug,
		"delete_volume": true,
	})
	recDel := httptest.NewRecorder()
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/v1/services/"+svc.ID, bytes.NewReader(delBody))
	for _, c := range cookies {
		reqDel.AddCookie(c)
	}
	mux.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusNoContent {
		t.Fatalf("delete: %d %s", recDel.Code, recDel.Body.String())
	}

	recGet := httptest.NewRecorder()
	reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/services/"+svc.ID, nil)
	for _, c := range cookies {
		reqGet.AddCookie(c)
	}
	mux.ServeHTTP(recGet, reqGet)
	if recGet.Code != http.StatusNotFound {
		t.Fatalf("get after delete: %d", recGet.Code)
	}
}
