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

func TestOwnerWithout2FABlockedFromProjects(t *testing.T) {
	t.Setenv("DOCKYARD_REQUIRE_OWNER_2FA", "1")
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

	loginBody, _ := json.Marshal(map[string]any{
		"email":    "owner@example.com",
		"password": "longpassword12",
		"remember": true,
	})
	recLogin := httptest.NewRecorder()
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	mux.ServeHTTP(recLogin, reqLogin)
	if recLogin.Code != http.StatusOK {
		t.Fatalf("login %d", recLogin.Code)
	}
	cookie := recLogin.Result().Cookies()[0]

	recProjects := httptest.NewRecorder()
	reqProjects := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	reqProjects.AddCookie(cookie)
	mux.ServeHTTP(recProjects, reqProjects)
	if recProjects.Code != http.StatusForbidden {
		t.Fatalf("projects status=%d want 403 until owner enables 2FA", recProjects.Code)
	}

	recTOTP := httptest.NewRecorder()
	reqTOTP := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/setup", bytes.NewReader([]byte("{}")))
	reqTOTP.AddCookie(cookie)
	mux.ServeHTTP(recTOTP, reqTOTP)
	if recTOTP.Code != http.StatusOK {
		t.Fatalf("totp setup should be allowed, got %d %s", recTOTP.Code, recTOTP.Body.String())
	}
}
