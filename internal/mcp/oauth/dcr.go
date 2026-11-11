package oauth

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

const dcrMaxBody = 32 << 10

type registeredClient struct {
	ClientID                  string
	RedirectURIs              []string
	TokenEndpointAuthMethod   string
	GrantTypes                []string
	ResponseTypes             []string
	IssuedAt                  int64
}

var (
	dcrMu      sync.RWMutex
	dcrClients = map[string]registeredClient{}
)

func lookupRegisteredClient(clientID string) (registeredClient, bool) {
	dcrMu.RLock()
	defer dcrMu.RUnlock()
	c, ok := dcrClients[clientID]
	return c, ok
}

func (c registeredClient) allowsRedirectURI(redirectURI, panelOrigin string) bool {
	for _, allowed := range c.RedirectURIs {
		if redirectURIMatches(allowed, redirectURI) && redirectURISafe(redirectURI) {
			return true
		}
	}
	return false
}

type dcrRequest struct {
	RedirectURIs            []string `json:"redirect_uris"`
	ClientName              string   `json:"client_name"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
}

func (s *Server) handleRegisterClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ct := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if ct != "" && !strings.HasPrefix(ct, "application/json") {
		writeDCRError(w, http.StatusUnsupportedMediaType, "invalid_client_metadata", "Content-Type must be application/json")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, dcrMaxBody))
	if err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	var req dcrRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeDCRError(w, http.StatusBadRequest, "invalid_client_metadata", "invalid JSON")
		return
	}
	if len(req.RedirectURIs) == 0 {
		writeDCRError(w, http.StatusBadRequest, "invalid_redirect_uri", "redirect_uris required")
		return
	}
	for _, ru := range req.RedirectURIs {
		if !redirectURISafe(ru) {
			writeDCRError(w, http.StatusBadRequest, "invalid_redirect_uri", "redirect URI not allowed")
			return
		}
	}
	authMethod := strings.TrimSpace(req.TokenEndpointAuthMethod)
	if authMethod == "" {
		authMethod = "none"
	}
	if authMethod != "none" {
		writeDCRError(w, http.StatusBadRequest, "invalid_client_metadata", "only public clients supported")
		return
	}
	grantTypes := req.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{"authorization_code"}
	}
	if !sliceContains(grantTypes, "authorization_code") {
		writeDCRError(w, http.StatusBadRequest, "invalid_client_metadata", "authorization_code grant required")
		return
	}
	responseTypes := req.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []string{"code"}
	}

	clientID := "mcp_" + ulid.Make().String()
	now := time.Now().Unix()
	rec := registeredClient{
		ClientID:                clientID,
		RedirectURIs:            append([]string(nil), req.RedirectURIs...),
		TokenEndpointAuthMethod: authMethod,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		IssuedAt:                now,
	}
	dcrMu.Lock()
	dcrClients[clientID] = rec
	dcrMu.Unlock()

	writeJSON(w, http.StatusCreated, map[string]any{
		"client_id":                   clientID,
		"client_id_issued_at":         now,
		"client_secret_expires_at":    0,
		"redirect_uris":               rec.RedirectURIs,
		"token_endpoint_auth_method":  authMethod,
		"grant_types":                 grantTypes,
		"response_types":              responseTypes,
	})
}

func writeDCRError(w http.ResponseWriter, status int, code, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":             code,
		"error_description": description,
	})
}
