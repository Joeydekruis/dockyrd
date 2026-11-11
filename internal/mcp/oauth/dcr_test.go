package oauth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDynamicClientRegistration(t *testing.T) {
	srv := &Server{}
	mux := http.NewServeMux()
	srv.Register(mux)

	body, _ := json.Marshal(map[string]any{
		"redirect_uris":              []string{"https://claude.ai/api/mcp/auth_callback"},
		"token_endpoint_auth_method": "none",
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/oauth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register status %d body %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	clientID, _ := resp["client_id"].(string)
	if clientID == "" {
		t.Fatal("missing client_id")
	}
	oc, err := ResolveOAuthClient(t.Context(), clientID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !oc.AllowsRedirectURI("https://claude.ai/api/mcp/auth_callback", "https://panel.example") {
		t.Fatal("DCR client should allow Claude redirect")
	}
}
