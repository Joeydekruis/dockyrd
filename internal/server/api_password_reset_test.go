package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestPasswordResetFlow(t *testing.T) {
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

	plain, hash, err := auth.NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(time.Hour).Unix()
	user, err := q.GetUserByEmail(context.Background(), "owner@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if err := q.InsertPasswordReset(context.Background(), sqlc.InsertPasswordResetParams{
		TokenHash: hash,
		UserID:    sqlNullUserID(user.ID),
		ExpiresAt: sql.NullInt64{Int64: expires, Valid: true},
	}); err != nil {
		t.Fatal(err)
	}

	forgotBody, _ := json.Marshal(map[string]string{"email": "owner@example.com"})
	recForgot := httptest.NewRecorder()
	reqForgot := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/forgot", bytes.NewReader(forgotBody))
	mux.ServeHTTP(recForgot, reqForgot)
	if recForgot.Code != http.StatusOK {
		t.Fatalf("forgot status=%d %s", recForgot.Code, recForgot.Body.String())
	}

	resetBody, _ := json.Marshal(map[string]string{
		"token":    plain,
		"password": "brandnewpassword1",
	})
	recReset := httptest.NewRecorder()
	reqReset := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset", bytes.NewReader(resetBody))
	mux.ServeHTTP(recReset, reqReset)
	if recReset.Code != http.StatusOK {
		t.Fatalf("reset status=%d %s", recReset.Code, recReset.Body.String())
	}

	loginBody, _ := json.Marshal(map[string]any{
		"email":    "owner@example.com",
		"password": "brandnewpassword1",
		"remember": true,
	})
	recLogin := httptest.NewRecorder()
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	mux.ServeHTTP(recLogin, reqLogin)
	if recLogin.Code != http.StatusOK {
		t.Fatalf("login after reset status=%d", recLogin.Code)
	}
}