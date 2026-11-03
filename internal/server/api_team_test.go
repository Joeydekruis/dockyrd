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

func TestTeamAPIRequiresAdmin(t *testing.T) {
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
	if recSetup.Code != http.StatusOK {
		t.Fatalf("setup status=%d", recSetup.Code)
	}
	cookies := recSetup.Result().Cookies()

	recUsers := httptest.NewRecorder()
	reqUsers := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	for _, c := range cookies {
		reqUsers.AddCookie(c)
	}
	mux.ServeHTTP(recUsers, reqUsers)
	if recUsers.Code != http.StatusOK {
		t.Fatalf("list users status=%d body=%s", recUsers.Code, recUsers.Body.String())
	}

	inviteBody, _ := json.Marshal(map[string]string{
		"email": "dev@example.com",
		"role":  "member",
	})
	recInvite := httptest.NewRecorder()
	reqInvite := httptest.NewRequest(http.MethodPost, "/api/v1/invites", bytes.NewReader(inviteBody))
	for _, c := range cookies {
		reqInvite.AddCookie(c)
	}
	mux.ServeHTTP(recInvite, reqInvite)
	if recInvite.Code != http.StatusCreated {
		t.Fatalf("invite status=%d body=%s", recInvite.Code, recInvite.Body.String())
	}

	recAudit := httptest.NewRecorder()
	reqAudit := httptest.NewRequest(http.MethodGet, "/api/v1/audit", nil)
	for _, c := range cookies {
		reqAudit.AddCookie(c)
	}
	mux.ServeHTTP(recAudit, reqAudit)
	if recAudit.Code != http.StatusOK {
		t.Fatalf("audit status=%d", recAudit.Code)
	}
}
