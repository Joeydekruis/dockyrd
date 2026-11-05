package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/apitoken"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestAPITokensCRUDAndWhoami(t *testing.T) {
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

	createBody, _ := json.Marshal(map[string]string{
		"name": "CI", "scope": apitoken.ScopeRead, "project_id": proj.ID,
	})
	recCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", bytes.NewReader(createBody))
	for _, c := range cookies {
		reqCreate.AddCookie(c)
	}
	mux.ServeHTTP(recCreate, reqCreate)
	if recCreate.Code != http.StatusCreated {
		t.Fatalf("create token status=%d %s", recCreate.Code, recCreate.Body.String())
	}
	var created struct {
		Token string `json:"token"`
		Entry struct {
			ID string `json:"id"`
		} `json:"entry"`
	}
	if err := json.NewDecoder(recCreate.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.Token == "" || created.Entry.ID == "" {
		t.Fatal("missing token response")
	}

	recMe := httptest.NewRecorder()
	reqMe := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+created.Token)
	mux.ServeHTTP(recMe, reqMe)
	if recMe.Code != http.StatusOK {
		t.Fatalf("whoami status=%d", recMe.Code)
	}

	recDel := httptest.NewRecorder()
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/v1/tokens/"+created.Entry.ID, nil)
	for _, c := range cookies {
		reqDel.AddCookie(c)
	}
	mux.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusNoContent {
		t.Fatalf("revoke status=%d", recDel.Code)
	}
}
