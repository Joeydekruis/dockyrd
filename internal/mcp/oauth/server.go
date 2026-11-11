package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/apitoken"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/instance"
	mcpscope "github.com/Joeydekruis/dockyrd/internal/mcp"
)

const accessTokenTTL = time.Hour

// SessionUser resolves the logged-in user from a dashboard session cookie.
type SessionUser func(r *http.Request) (sqlc.GetUserByIDRow, bool)

// Server serves MCP OAuth metadata and token endpoints on the panel origin.
type Server struct {
	Cfg         config.Config
	Q           *sqlc.Queries
	SessionUser SessionUser
	LoadPanel   func(ctx context.Context) (instance.Panel, error)
	AllowAPIToken func(ctx context.Context) bool
}

func (s *Server) origin(r *http.Request) string {
	panel, _ := s.LoadPanel(r.Context())
	return PanelOrigin(r, s.Cfg, panel)
}

func (s *Server) resourceURL(r *http.Request) string {
	return s.origin(r) + "/mcp"
}

func (s *Server) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /.well-known/oauth-protected-resource", s.handleProtectedResourceMetadata)
	mux.HandleFunc("GET /.well-known/oauth-protected-resource/mcp", s.handleProtectedResourceMetadata)
	mux.HandleFunc("GET /.well-known/oauth-authorization-server", s.handleAuthorizationServerMetadata)
	mux.HandleFunc("GET /.well-known/openid-configuration", s.handleAuthorizationServerMetadata)
	mux.HandleFunc("GET /oauth/authorize", s.handleAuthorize)
	mux.HandleFunc("POST /oauth/token", s.handleToken)
	mux.HandleFunc("POST /oauth/register", s.handleRegisterClient)
}

func (s *Server) handleProtectedResourceMetadata(w http.ResponseWriter, r *http.Request) {
	origin := s.origin(r)
	writeJSON(w, http.StatusOK, map[string]any{
		"resource":                s.resourceURL(r),
		"authorization_servers":   []string{origin},
		"scopes_supported":        mcpscope.AllScopes,
		"bearer_methods_supported": []string{"header"},
	})
}

func (s *Server) handleAuthorizationServerMetadata(w http.ResponseWriter, r *http.Request) {
	origin := s.origin(r)
	meta := map[string]any{
		"issuer":                                origin,
		"authorization_endpoint":                origin + "/oauth/authorize",
		"token_endpoint":                        origin + "/oauth/token",
		"registration_endpoint":                 origin + "/oauth/register",
		"scopes_supported":                      mcpscope.AllScopes,
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code"},
		"token_endpoint_auth_methods_supported": []string{"none"},
		"code_challenge_methods_supported":      []string{"S256"},
		"client_id_metadata_document_supported": true,
	}
	writeJSON(w, http.StatusOK, meta)
}

func (s *Server) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Get("response_type") != "code" {
		http.Error(w, "unsupported_response_type", http.StatusBadRequest)
		return
	}
	redirectURI := strings.TrimSpace(q.Get("redirect_uri"))
	if redirectURI == "" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	origin := s.origin(r)
	if q.Get("code_challenge_method") != "S256" || strings.TrimSpace(q.Get("code_challenge")) == "" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	clientID := strings.TrimSpace(q.Get("client_id"))
	if clientID == "" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	oauthClient, err := ResolveOAuthClient(r.Context(), clientID, http.DefaultClient)
	if err != nil {
		http.Error(w, "invalid_client", http.StatusBadRequest)
		return
	}
	if !oauthClient.AllowsRedirectURI(redirectURI, origin) {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	scopes := mcpscope.ParseScopeList(q.Get("scope"))
	if len(scopes) == 0 {
		scopes = []string{mcpscope.ScopeRead}
	}

	user, ok := s.SessionUser(r)
	if !ok {
		loginURL := "/sign-in?return=" + url.QueryEscape(r.URL.Path+"?"+r.URL.RawQuery)
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}
	if !sessionAllowsMCScopes(user, scopes) {
		http.Error(w, "invalid_scope", http.StatusBadRequest)
		return
	}
	if q.Get("consent") != "1" {
		renderOAuthConsent(w, r, oauthClient.DisplayID(), scopes)
		return
	}

	codePlain, codeHash, err := randomToken()
	if err != nil {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}
	expires := time.Now().Add(10 * time.Minute).Unix()
	err = s.Q.InsertMCPOAuthCode(r.Context(), sqlc.InsertMCPOAuthCodeParams{
		CodeHash:       codeHash,
		ClientID:       clientID,
		UserID:         user.ID,
		RedirectUri:    redirectURI,
		CodeChallenge:  q.Get("code_challenge"),
		Scope:          strings.Join(scopes, " "),
		ExpiresAt:      expires,
	})
	if err != nil {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}
	_ = codePlain // codePlain used below

	redir, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	qv := redir.Query()
	qv.Set("code", codePlain)
	if state := q.Get("state"); state != "" {
		qv.Set("state", state)
	}
	redir.RawQuery = qv.Encode()
	http.Redirect(w, r, redir.String(), http.StatusFound)
}

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	if r.Form.Get("grant_type") != "authorization_code" {
		http.Error(w, "unsupported_grant_type", http.StatusBadRequest)
		return
	}
	codePlain := strings.TrimSpace(r.Form.Get("code"))
	codeVerifier := strings.TrimSpace(r.Form.Get("code_verifier"))
	if codePlain == "" || codeVerifier == "" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	sum := sha256.Sum256([]byte(codePlain))
	codeHash := sum[:]

	row, err := s.Q.GetMCPOAuthCode(r.Context(), codeHash)
	if err != nil {
		writeOAuthJSONError(w, http.StatusBadRequest, "invalid_grant", "invalid authorization code")
		return
	}
	_ = s.Q.DeleteMCPOAuthCode(r.Context(), codeHash)
	if time.Now().Unix() > row.ExpiresAt {
		writeOAuthJSONError(w, http.StatusBadRequest, "invalid_grant", "authorization code expired")
		return
	}
	clientID := strings.TrimSpace(r.Form.Get("client_id"))
	if clientID != "" && clientID != row.ClientID {
		writeOAuthJSONError(w, http.StatusBadRequest, "invalid_grant", "client_id mismatch")
		return
	}
	redirectURI := strings.TrimSpace(r.Form.Get("redirect_uri"))
	if redirectURI != "" && !redirectURIMatchesAuth(row.RedirectUri, redirectURI) {
		writeOAuthJSONError(w, http.StatusBadRequest, "invalid_grant", "redirect_uri mismatch")
		return
	}
	if redirectURI == "" && row.RedirectUri != "" {
		writeOAuthJSONError(w, http.StatusBadRequest, "invalid_grant", "redirect_uri required")
		return
	}
	if !pkceValid(row.CodeChallenge, codeVerifier) {
		writeOAuthJSONError(w, http.StatusBadRequest, "invalid_grant", "invalid code_verifier")
		return
	}

	accessPlain, accessHash, err := randomToken()
	if err != nil {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}
	now := time.Now().Unix()
	expires := time.Now().Add(accessTokenTTL).Unix()
	if err := s.Q.InsertMCPAccessToken(r.Context(), sqlc.InsertMCPAccessTokenParams{
		TokenHash:  accessHash,
		UserID:     row.UserID,
		Scope:      row.Scope,
		ExpiresAt:  expires,
		CreatedAt:  now,
	}); err != nil {
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": accessPlain,
		"token_type":   "Bearer",
		"expires_in":   int(accessTokenTTL.Seconds()),
		"scope":        row.Scope,
	})
}

func pkceValid(challenge, verifier string) bool {
	h := sha256.Sum256([]byte(verifier))
	got := base64.RawURLEncoding.EncodeToString(h[:])
	return got == challenge
}

func randomToken() (plain string, hash []byte, err error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", nil, err
	}
	plain = base64.RawURLEncoding.EncodeToString(b[:])
	sum := sha256.Sum256([]byte(plain))
	return plain, sum[:], nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func sessionAllowsMCScopes(user sqlc.GetUserByIDRow, scopes []string) bool {
	for _, sc := range scopes {
		switch sc {
		case mcpscope.ScopeRead:
			continue
		case mcpscope.ScopeDeploy:
			if user.Role == "viewer" {
				return false
			}
		case mcpscope.ScopeWrite:
			if user.Role != "owner" && user.Role != "admin" && user.Role != "member" {
				return false
			}
		case mcpscope.ScopeInstRO:
			if user.Role != "owner" && user.Role != "admin" {
				return false
			}
		}
	}
	return true
}

// VerifyBearerToken validates Bearer tokens for MCP HTTP. apiTok is set when a dyk_* token was used.
func (s *Server) VerifyBearerToken(ctx context.Context, token string, r *http.Request) (scopes []string, userID string, exp time.Time, apiTok *sqlc.ApiToken, err error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, "", time.Time{}, nil, errors.New("empty token")
	}

	// OAuth access token
	sum := sha256.Sum256([]byte(token))
	row, err := s.Q.GetMCPAccessTokenByHash(ctx, sum[:])
	if err == nil {
		if time.Now().Unix() > row.ExpiresAt {
			return nil, "", time.Time{}, nil, errors.New("expired")
		}
		return strings.Fields(row.Scope), row.UserID, time.Unix(row.ExpiresAt, 0), nil, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, "", time.Time{}, nil, err
	}

	// Optional API token bridge
	if s.AllowAPIToken == nil || !s.AllowAPIToken(ctx) {
		return nil, "", time.Time{}, nil, errors.New("invalid token")
	}
	if !strings.HasPrefix(token, apitoken.PlainPrefix) {
		return nil, "", time.Time{}, nil, errors.New("invalid token")
	}
	hash, err := apitoken.Hash(token)
	if err != nil {
		return nil, "", time.Time{}, nil, errors.New("invalid token")
	}
	apiRow, err := s.Q.GetAPITokenByHash(ctx, hash)
	if err != nil {
		return nil, "", time.Time{}, nil, errors.New("invalid token")
	}
	if !apitoken.IsUserScope(apiRow.Scope) || apiRow.RevokedAt.Valid {
		return nil, "", time.Time{}, nil, errors.New("invalid token")
	}
	if apiRow.ExpiresAt.Valid && time.Now().Unix() > apiRow.ExpiresAt.Int64 {
		return nil, "", time.Time{}, nil, errors.New("invalid token")
	}
	if !apiRow.CreatedBy.Valid {
		return nil, "", time.Time{}, nil, errors.New("invalid token")
	}
	mcpScopes := apiTokenToMCPScopes(apiRow.Scope)
	tokenExp := time.Now().Add(24 * time.Hour)
	if apiRow.ExpiresAt.Valid {
		tokenExp = time.Unix(apiRow.ExpiresAt.Int64, 0)
	}
	return mcpScopes, apiRow.CreatedBy.String, tokenExp, &apiRow, nil
}

func apiTokenToMCPScopes(scope string) []string {
	var out []string
	if apitoken.ScopeAllows(scope, apitoken.ScopeRead) {
		out = append(out, mcpscope.ScopeRead)
	}
	if apitoken.ScopeAllows(scope, apitoken.ScopeDeploy) {
		out = append(out, mcpscope.ScopeDeploy)
	}
	if apitoken.ScopeAllows(scope, apitoken.ScopeWrite) {
		out = append(out, mcpscope.ScopeWrite)
	}
	return out
}

// CleanupExpiredTokens removes stale MCP access tokens (best-effort).
func (s *Server) CleanupExpiredTokens(ctx context.Context) {
	if s.Q == nil {
		return
	}
	_ = s.Q.DeleteExpiredMCPAccessTokens(ctx, time.Now().Unix())
}

// MintTestAccessToken is used in tests to issue a token without the OAuth UI flow.
func (s *Server) MintTestAccessToken(ctx context.Context, userID, scope string) (plain string, err error) {
	if s.Q == nil {
		return "", fmt.Errorf("no db")
	}
	plain, hash, err := randomToken()
	if err != nil {
		return "", err
	}
	now := time.Now().Unix()
	exp := time.Now().Add(accessTokenTTL).Unix()
	err = s.Q.InsertMCPAccessToken(ctx, sqlc.InsertMCPAccessTokenParams{
		TokenHash:  hash,
		UserID:     userID,
		Scope:      scope,
		ExpiresAt:  exp,
		CreatedAt:  now,
	})
	if err != nil {
		return "", err
	}
	_ = ulid.Make()
	return plain, nil
}
