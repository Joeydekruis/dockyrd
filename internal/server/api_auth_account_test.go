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

func TestPatchAccountAndNotifications(t *testing.T) {
	mux, cookie := testMuxWithOwnerSession(t)

	body, _ := json.Marshal(map[string]string{
		"name":  "Updated Name",
		"email": "updated@example.com",
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/auth/account", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("patch account: %d %s", w.Code, w.Body.String())
	}

	reqN := httptest.NewRequest(http.MethodGet, "/api/v1/auth/notifications", nil)
	reqN.AddCookie(cookie)
	wN := httptest.NewRecorder()
	mux.ServeHTTP(wN, reqN)
	if wN.Code != http.StatusOK {
		t.Fatalf("get notifications: %d", wN.Code)
	}

	patchBody, _ := json.Marshal(map[string]bool{
		"deployments":     false,
		"security_alerts": true,
		"product_updates": true,
	})
	reqP := httptest.NewRequest(http.MethodPatch, "/api/v1/auth/notifications", bytes.NewReader(patchBody))
	reqP.Header.Set("Content-Type", "application/json")
	reqP.AddCookie(cookie)
	wP := httptest.NewRecorder()
	mux.ServeHTTP(wP, reqP)
	if wP.Code != http.StatusOK {
		t.Fatalf("patch notifications: %d %s", wP.Code, wP.Body.String())
	}

	reqL := httptest.NewRequest(http.MethodGet, "/api/v1/auth/locale", nil)
	reqL.AddCookie(cookie)
	wL := httptest.NewRecorder()
	mux.ServeHTTP(wL, reqL)
	if wL.Code != http.StatusOK {
		t.Fatalf("get locale: %d", wL.Code)
	}
	var defaultLocale localePrefs
	if err := json.Unmarshal(wL.Body.Bytes(), &defaultLocale); err != nil {
		t.Fatal(err)
	}
	if defaultLocale.Timezone != "UTC" || defaultLocale.TimeFormat != "24h" {
		t.Fatalf("unexpected default locale: %+v", defaultLocale)
	}

	patchLocale, _ := json.Marshal(map[string]string{
		"timezone":    "Europe/Amsterdam",
		"time_format": "12h",
	})
	reqLP := httptest.NewRequest(http.MethodPatch, "/api/v1/auth/locale", bytes.NewReader(patchLocale))
	reqLP.Header.Set("Content-Type", "application/json")
	reqLP.AddCookie(cookie)
	wLP := httptest.NewRecorder()
	mux.ServeHTTP(wLP, reqLP)
	if wLP.Code != http.StatusOK {
		t.Fatalf("patch locale: %d %s", wLP.Code, wLP.Body.String())
	}

	reqL2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/locale", nil)
	reqL2.AddCookie(cookie)
	wL2 := httptest.NewRecorder()
	mux.ServeHTTP(wL2, reqL2)
	if wL2.Code != http.StatusOK {
		t.Fatalf("get locale after patch: %d", wL2.Code)
	}
	var storedLocale localePrefs
	if err := json.Unmarshal(wL2.Body.Bytes(), &storedLocale); err != nil {
		t.Fatal(err)
	}
	if storedLocale.Timezone != "Europe/Amsterdam" || storedLocale.TimeFormat != "12h" {
		t.Fatalf("expected persisted locale, got %+v", storedLocale)
	}

	reqBad := httptest.NewRequest(http.MethodPatch, "/api/v1/auth/locale", bytes.NewReader([]byte(`{"timezone":"Invalid/Zone","time_format":"24h"}`)))
	reqBad.Header.Set("Content-Type", "application/json")
	reqBad.AddCookie(cookie)
	wBad := httptest.NewRecorder()
	mux.ServeHTTP(wBad, reqBad)
	if wBad.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad tz, got %d", wBad.Code)
	}
}

func TestPasswordChangeWrongCurrent(t *testing.T) {
	mux, cookie := testMuxWithOwnerSession(t)
	body, _ := json.Marshal(map[string]string{
		"current_password": "wrong",
		"new_password":     "brandnewpassword1",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/change", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func testMuxWithOwnerSession(t *testing.T) (*http.ServeMux, *http.Cookie) {
	t.Helper()
	t.Setenv("dockyrd_REQUIRE_OWNER_2FA", "0")
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
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody))
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("setup status=%d", rec.Code)
	}

	loginBody, _ := json.Marshal(map[string]string{
		"email":    "owner@example.com",
		"password": "longpassword12",
	})
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("login status=%d", rec2.Code)
	}
	cookies := rec2.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("missing session cookie")
	}
	return mux, cookies[0]
}
