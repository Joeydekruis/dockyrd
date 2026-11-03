package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestInviteAcceptFlow(t *testing.T) {
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
	deps := apiDeps{cfg: cfg, q: q, audit: audit.Logger{Q: q}}
	registerAPIRoutes(mux, deps)

	setupBody, _ := json.Marshal(map[string]string{
		"token":    initRes.Token,
		"name":     "Owner",
		"email":    "owner@example.com",
		"password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	reqSetup := httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody))
	mux.ServeHTTP(recSetup, reqSetup)
	ownerCookies := recSetup.Result().Cookies()

	inviteBody, _ := json.Marshal(map[string]string{
		"email": "member@example.com",
		"role":  "member",
	})
	recInvite := httptest.NewRecorder()
	reqInvite := httptest.NewRequest(http.MethodPost, "/api/v1/invites", bytes.NewReader(inviteBody))
	for _, c := range ownerCookies {
		reqInvite.AddCookie(c)
	}
	mux.ServeHTTP(recInvite, reqInvite)
	if recInvite.Code != http.StatusCreated {
		t.Fatalf("invite status=%d body=%s", recInvite.Code, recInvite.Body.String())
	}
	var created struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(recInvite.Body).Decode(&created); err != nil || created.Token == "" {
		t.Fatalf("invite token missing: %v", err)
	}

	recPreview := httptest.NewRecorder()
	reqPreview := httptest.NewRequest(http.MethodGet, "/api/v1/invites/preview?token="+created.Token, nil)
	mux.ServeHTTP(recPreview, reqPreview)
	if recPreview.Code != http.StatusOK {
		t.Fatalf("preview status=%d", recPreview.Code)
	}

	acceptBody, _ := json.Marshal(map[string]string{
		"token":    created.Token,
		"name":     "Member",
		"password": "anotherpassword1",
	})
	recAccept := httptest.NewRecorder()
	reqAccept := httptest.NewRequest(http.MethodPost, "/api/v1/invites/accept", bytes.NewReader(acceptBody))
	mux.ServeHTTP(recAccept, reqAccept)
	if recAccept.Code != http.StatusOK {
		t.Fatalf("accept status=%d body=%s", recAccept.Code, recAccept.Body.String())
	}
	memberCookies := recAccept.Result().Cookies()
	if len(memberCookies) == 0 {
		t.Fatal("expected session cookie after accept")
	}

	recMe := httptest.NewRecorder()
	reqMe := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	for _, c := range memberCookies {
		reqMe.AddCookie(c)
	}
	mux.ServeHTTP(recMe, reqMe)
	if recMe.Code != http.StatusOK {
		t.Fatalf("me status=%d", recMe.Code)
	}
}

func TestSearchReturnsProjects(t *testing.T) {
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
	deps := apiDeps{cfg: cfg, q: q, audit: audit.Logger{Q: q}}
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
	cookies := recSetup.Result().Cookies()

	createBody, _ := json.Marshal(map[string]string{"name": "Acme Corp"})
	recCreate := httptest.NewRecorder()
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(createBody))
	for _, c := range cookies {
		reqCreate.AddCookie(c)
	}
	mux.ServeHTTP(recCreate, reqCreate)
	if recCreate.Code != http.StatusCreated {
		t.Fatalf("project create %d", recCreate.Code)
	}

	recSearch := httptest.NewRecorder()
	reqSearch := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=acme", nil)
	for _, c := range cookies {
		reqSearch.AddCookie(c)
	}
	mux.ServeHTTP(recSearch, reqSearch)
	if recSearch.Code != http.StatusOK {
		t.Fatalf("search status=%d", recSearch.Code)
	}
	var payload struct {
		Results []searchHit `json:"results"`
	}
	if err := json.NewDecoder(recSearch.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, hit := range payload.Results {
		if hit.Kind == "project" && hit.Label == "Acme Corp" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected project in search: %+v", payload.Results)
	}
}
