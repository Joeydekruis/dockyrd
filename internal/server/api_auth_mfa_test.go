package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestLoginMFAFlow(t *testing.T) {
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
	cookies := recSetup.Result().Cookies()

	setupTotpBody, _ := json.Marshal(map[string]string{})
	recTSetup := httptest.NewRecorder()
	reqTSetup := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/setup", bytes.NewReader(setupTotpBody))
	for _, c := range cookies {
		reqTSetup.AddCookie(c)
	}
	mux.ServeHTTP(recTSetup, reqTSetup)
	if recTSetup.Code != http.StatusOK {
		t.Fatalf("totp setup %d %s", recTSetup.Code, recTSetup.Body.String())
	}
	var pending struct {
		Secret string `json:"secret"`
	}
	_ = json.NewDecoder(recTSetup.Body).Decode(&pending)
	code, err := totp.GenerateCode(pending.Secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	enableBody, _ := json.Marshal(map[string]string{"secret": pending.Secret, "code": code})
	recEnable := httptest.NewRecorder()
	reqEnable := httptest.NewRequest(http.MethodPost, "/api/v1/auth/totp/enable", bytes.NewReader(enableBody))
	for _, c := range cookies {
		reqEnable.AddCookie(c)
	}
	mux.ServeHTTP(recEnable, reqEnable)
	if recEnable.Code != http.StatusOK {
		t.Fatalf("totp enable %d %s", recEnable.Code, recEnable.Body.String())
	}

	loginBody, _ := json.Marshal(map[string]any{
		"email":    "owner@example.com",
		"password": "longpassword12",
		"remember": true,
	})
	recLogin := httptest.NewRecorder()
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	mux.ServeHTTP(recLogin, reqLogin)
	if recLogin.Code != http.StatusAccepted {
		t.Fatalf("login status=%d want 202", recLogin.Code)
	}
	loginCookies := recLogin.Result().Cookies()

	recMeBlocked := httptest.NewRecorder()
	reqMeBlocked := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	for _, c := range loginCookies {
		reqMeBlocked.AddCookie(c)
	}
	mux.ServeHTTP(recMeBlocked, reqMeBlocked)
	if recMeBlocked.Code != http.StatusUnauthorized {
		t.Fatalf("me before mfa should 401, got %d", recMeBlocked.Code)
	}
	var meBlocked map[string]any
	_ = json.NewDecoder(recMeBlocked.Body).Decode(&meBlocked)
	if meBlocked["mfa_required"] != true {
		t.Fatalf("me before mfa should set mfa_required, got %#v", meBlocked)
	}

	code2, _ := totp.GenerateCode(pending.Secret, time.Now().UTC())
	mfaBody, _ := json.Marshal(map[string]string{"code": code2})
	recMFA := httptest.NewRecorder()
	reqMFA := httptest.NewRequest(http.MethodPost, "/api/v1/auth/mfa", bytes.NewReader(mfaBody))
	for _, c := range loginCookies {
		reqMFA.AddCookie(c)
	}
	mux.ServeHTTP(recMFA, reqMFA)
	if recMFA.Code != http.StatusOK {
		t.Fatalf("mfa status=%d body=%s", recMFA.Code, recMFA.Body.String())
	}

	recMe := httptest.NewRecorder()
	reqMe := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	for _, c := range loginCookies {
		reqMe.AddCookie(c)
	}
	mux.ServeHTTP(recMe, reqMe)
	if recMe.Code != http.StatusOK {
		t.Fatalf("me after mfa status=%d", recMe.Code)
	}
}
