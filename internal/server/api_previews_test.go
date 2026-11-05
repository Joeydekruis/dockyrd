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

func TestPreviewConfigPatch(t *testing.T) {
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
	cookies := recSetup.Result().Cookies()

	projBody, _ := json.Marshal(map[string]string{"name": "Demo"})
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
		"kind":       "app",
		"name":       "API",
		"config":     map[string]any{},
	})
	recSvc := httptest.NewRecorder()
	reqSvc := httptest.NewRequest(http.MethodPost, "/api/v1/services", bytes.NewReader(svcBody))
	for _, c := range cookies {
		reqSvc.AddCookie(c)
	}
	mux.ServeHTTP(recSvc, reqSvc)
	var svc serviceResponse
	_ = json.NewDecoder(recSvc.Body).Decode(&svc)

	patchBody, _ := json.Marshal(map[string]bool{"enabled": true})
	recPatch := httptest.NewRecorder()
	reqPatch := httptest.NewRequest(http.MethodPatch, "/api/v1/services/"+svc.ID+"/preview-config", bytes.NewReader(patchBody))
	for _, c := range cookies {
		reqPatch.AddCookie(c)
	}
	mux.ServeHTTP(recPatch, reqPatch)
	if recPatch.Code != http.StatusOK {
		t.Fatalf("patch preview-config status=%d %s", recPatch.Code, recPatch.Body.String())
	}

	recList := httptest.NewRecorder()
	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/services/"+svc.ID+"/previews", nil)
	for _, c := range cookies {
		reqList.AddCookie(c)
	}
	mux.ServeHTTP(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("list previews status=%d", recList.Code)
	}
}
