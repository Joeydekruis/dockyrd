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

func TestProjectsCRUD(t *testing.T) {
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
		"token":    initRes.Token,
		"name":     "Joey",
		"email":    "owner@example.com",
		"password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	reqSetup := httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody))
	mux.ServeHTTP(recSetup, reqSetup)
	if recSetup.Code != http.StatusOK {
		t.Fatalf("setup status=%d body=%s", recSetup.Code, recSetup.Body.String())
	}

	srv, err := q.GetLocalServer(t.Context())
	if err != nil {
		t.Fatalf("local server: %v", err)
	}
	if srv.ID == "" {
		t.Fatal("expected local server id")
	}

	sessionCookies := recSetup.Result().Cookies()

	createBody, _ := json.Marshal(map[string]string{"name": "Client X", "description": "Demo"})
	recCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(createBody))
	for _, c := range sessionCookies {
		reqCreate.AddCookie(c)
	}
	mux.ServeHTTP(recCreate, reqCreate)
	if recCreate.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", recCreate.Code, recCreate.Body.String())
	}
	var created projectResponse
	if err := json.NewDecoder(recCreate.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.Slug != "client-x" || created.ServerID != srv.ID {
		t.Fatalf("unexpected project: %+v", created)
	}

	recList := httptest.NewRecorder()
	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	for _, c := range sessionCookies {
		reqList.AddCookie(c)
	}
	mux.ServeHTTP(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("list status=%d", recList.Code)
	}

	patchBody, _ := json.Marshal(map[string]string{"name": "Client Y"})
	recPatch := httptest.NewRecorder()
	reqPatch := httptest.NewRequest(http.MethodPatch, "/api/v1/projects/"+created.ID, bytes.NewReader(patchBody))
	for _, c := range sessionCookies {
		reqPatch.AddCookie(c)
	}
	mux.ServeHTTP(recPatch, reqPatch)
	if recPatch.Code != http.StatusOK {
		t.Fatalf("patch status=%d body=%s", recPatch.Code, recPatch.Body.String())
	}

	recDel := httptest.NewRecorder()
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+created.ID, nil)
	for _, c := range sessionCookies {
		reqDel.AddCookie(c)
	}
	mux.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusNoContent {
		t.Fatalf("delete status=%d", recDel.Code)
	}

	recAnon := httptest.NewRecorder()
	reqAnon := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	mux.ServeHTTP(recAnon, reqAnon)
	if recAnon.Code != http.StatusUnauthorized {
		t.Fatalf("anon list status=%d", recAnon.Code)
	}
}
