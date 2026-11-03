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

func TestDeploymentEnqueue(t *testing.T) {
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
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))
	if recSetup.Code != http.StatusOK {
		t.Fatalf("setup %d", recSetup.Code)
	}

	sessionCookies := recSetup.Result().Cookies()

	createProj, _ := json.Marshal(map[string]string{"name": "Demo"})
	recProj := httptest.NewRecorder()
	reqProj := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(createProj))
	for _, c := range sessionCookies {
		reqProj.AddCookie(c)
	}
	mux.ServeHTTP(recProj, reqProj)
	var proj struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(recProj.Body.Bytes(), &proj)

	createSvc, _ := json.Marshal(map[string]any{
		"project_id": proj.ID, "name": "web", "kind": "app",
	})
	recSvc := httptest.NewRecorder()
	reqSvc := httptest.NewRequest(http.MethodPost, "/api/v1/services", bytes.NewReader(createSvc))
	for _, c := range sessionCookies {
		reqSvc.AddCookie(c)
	}
	mux.ServeHTTP(recSvc, reqSvc)
	var svc struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(recSvc.Body.Bytes(), &svc)

	recDep := httptest.NewRecorder()
	reqDep := httptest.NewRequest(http.MethodPost, "/api/v1/services/"+svc.ID+"/deployments", nil)
	for _, c := range sessionCookies {
		reqDep.AddCookie(c)
	}
	mux.ServeHTTP(recDep, reqDep)
	if recDep.Code != http.StatusCreated {
		t.Fatalf("deploy status=%d body=%s", recDep.Code, recDep.Body.String())
	}

	recList := httptest.NewRecorder()
	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/services/"+svc.ID+"/deployments", nil)
	for _, c := range sessionCookies {
		reqList.AddCookie(c)
	}
	mux.ServeHTTP(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("list deployments %d", recList.Code)
	}
}
