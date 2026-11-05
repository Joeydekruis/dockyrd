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

func TestBearerTokenListsProjectServices(t *testing.T) {
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
		"token": initRes.Token, "name": "Owner", "email": "o@example.com", "password": "longpassword12",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody))
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("setup=%d", rec.Code)
	}

	loginBody, _ := json.Marshal(map[string]string{"email": "o@example.com", "password": "longpassword12"})
	recLogin := httptest.NewRecorder()
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	for _, c := range rec.Result().Cookies() {
		reqLogin.AddCookie(c)
	}
	mux.ServeHTTP(recLogin, reqLogin)

	meRec := httptest.NewRecorder()
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	for _, c := range recLogin.Result().Cookies() {
		meReq.AddCookie(c)
	}
	mux.ServeHTTP(meRec, meReq)
	var me struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(meRec.Body.Bytes(), &me)

	projBody, _ := json.Marshal(map[string]string{"name": "Demo"})
	recProj := httptest.NewRecorder()
	reqProj := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(projBody))
	for _, c := range recLogin.Result().Cookies() {
		reqProj.AddCookie(c)
	}
	mux.ServeHTTP(recProj, reqProj)
	var proj projectResponse
	_ = json.NewDecoder(recProj.Body).Decode(&proj)
	projectID := proj.ID

	createTok, _ := json.Marshal(map[string]string{
		"name": "ci", "scope": apitoken.ScopeRead, "project_id": projectID,
	})
	tokRec := httptest.NewRecorder()
	tokReq := httptest.NewRequest(http.MethodPost, "/api/v1/tokens", bytes.NewReader(createTok))
	for _, c := range recLogin.Result().Cookies() {
		tokReq.AddCookie(c)
	}
	mux.ServeHTTP(tokRec, tokReq)
	if tokRec.Code != http.StatusCreated {
		t.Fatalf("create token=%d %s", tokRec.Code, tokRec.Body.String())
	}
	var tokResp struct {
		Token string `json:"token"`
	}
	_ = json.Unmarshal(tokRec.Body.Bytes(), &tokResp)
	if tokResp.Token == "" {
		t.Fatal("missing plaintext token")
	}

	svcRec := httptest.NewRecorder()
	svcReq := httptest.NewRequest(http.MethodGet, "/api/v1/services?project_id="+projectID, nil)
	svcReq.Header.Set("Authorization", "Bearer "+tokResp.Token)
	mux.ServeHTTP(svcRec, svcReq)
	if svcRec.Code != http.StatusOK {
		t.Fatalf("list services bearer=%d %s", svcRec.Code, svcRec.Body.String())
	}

	// Token must not access token admin routes.
	adminRec := httptest.NewRecorder()
	adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/tokens", nil)
	adminReq.Header.Set("Authorization", "Bearer "+tokResp.Token)
	mux.ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusForbidden {
		t.Fatalf("tokens admin bearer=%d want 403", adminRec.Code)
	}
}
