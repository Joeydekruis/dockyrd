package server

import (
	"context"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	mcpsoaut "github.com/Joeydekruis/dockyrd/internal/mcp/oauth"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestLoginRedirectsToSignIn(t *testing.T) {
	mux := http.NewServeMux()
	registerAPIRoutes(mux, apiDeps{})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/login?return=%2Foauth%2Fauthorize", nil))
	if rec.Code != http.StatusFound {
		t.Fatalf("status %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "/sign-in") {
		t.Fatalf("location %q", loc)
	}
}

func TestMCPProtectedResourceMetadata(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	mux := http.NewServeMux()
	registerAPIRoutes(mux, apiDeps{cfg: cfg, q: q})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/.well-known/oauth-protected-resource", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if res, ok := body["resource"].(string); !ok || res == "" {
		t.Fatalf("missing resource: %#v", body)
	}
	if _, ok := body["scopes_supported"]; !ok {
		t.Fatal("missing scopes_supported")
	}
}

func TestMCPOAuthPKCERoundtrip(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
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
		"token": initRes.Token, "name": "Joey", "email": "mcp@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))
	cookies := recSetup.Result().Cookies()

	verifier := "test-verifier-123456789"
	ch := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(ch[:])

	authQ := url.Values{}
	authQ.Set("response_type", "code")
	authQ.Set("client_id", "cursor")
	authQ.Set("redirect_uri", "http://127.0.0.1/callback")
	authQ.Set("code_challenge", challenge)
	authQ.Set("code_challenge_method", "S256")
	authQ.Set("scope", "mcp:read mcp:deploy")
	authQ.Set("state", "st")
	authQ.Set("consent", "1")

	recAuth := httptest.NewRecorder()
	reqAuth := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+authQ.Encode(), nil)
	for _, c := range cookies {
		reqAuth.AddCookie(c)
	}
	mux.ServeHTTP(recAuth, reqAuth)
	if recAuth.Code != http.StatusFound {
		t.Fatalf("authorize status %d body %s", recAuth.Code, recAuth.Body.String())
	}
	loc, err := url.Parse(recAuth.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	code := loc.Query().Get("code")
	if code == "" {
		t.Fatal("missing code in redirect")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("code_verifier", verifier)
	recTok := httptest.NewRecorder()
	reqTok := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	reqTok.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	mux.ServeHTTP(recTok, reqTok)
	if recTok.Code != http.StatusOK {
		t.Fatalf("token status %d %s", recTok.Code, recTok.Body.String())
	}
	var tok map[string]any
	if err := json.NewDecoder(recTok.Body).Decode(&tok); err != nil {
		t.Fatal(err)
	}
	access, _ := tok["access_token"].(string)
	if access == "" {
		t.Fatal("no access_token")
	}

	oauthSrv := &mcpsoaut.Server{Cfg: cfg, Q: q, AllowAPIToken: func(context.Context) bool { return false }}
	scopes, userID, _, _, err := oauthSrv.VerifyBearerToken(t.Context(), access, httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil || userID == "" || len(scopes) == 0 {
		t.Fatalf("verify: user=%q scopes=%v err=%v", userID, scopes, err)
	}
}

func TestMCPPostUnauthorized(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	mux := http.NewServeMux()
	registerAPIRoutes(mux, apiDeps{cfg: cfg, q: q})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d want 401", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("WWW-Authenticate"), "resource_metadata=") {
		t.Fatalf("missing WWW-Authenticate: %q", rec.Header().Get("WWW-Authenticate"))
	}
}
