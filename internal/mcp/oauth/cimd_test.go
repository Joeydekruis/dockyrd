package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectURIMatchesLoopbackPort(t *testing.T) {
	if !redirectURIMatches("http://127.0.0.1/callback", "http://127.0.0.1:53682/callback") {
		t.Fatal("expected loopback port-insensitive match")
	}
	if redirectURIMatches("https://claude.ai/api/mcp/auth_callback", "https://evil.example/callback") {
		t.Fatal("should not match different host")
	}
}

func TestResolveOAuthClientLegacy(t *testing.T) {
	c, err := ResolveOAuthClient(t.Context(), "cursor", nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.LegacyID != "cursor" || c.Metadata != nil {
		t.Fatalf("%+v", c)
	}
	if !c.AllowsRedirectURI("http://127.0.0.1/callback", "") {
		t.Fatal("legacy loopback redirect denied")
	}
}

func TestResolveOAuthClientCIMD(t *testing.T) {
	var docURL string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/client.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"client_id": "` + docURL + `",
			"client_name": "Test",
			"redirect_uris": ["http://localhost/callback"],
			"grant_types": ["authorization_code"],
			"response_types": ["code"],
			"token_endpoint_auth_method": "none"
		}`))
	}))
	t.Cleanup(srv.Close)
	docURL = srv.URL + "/client.json"

	client := srv.Client()
	clientID := docURL

	c, err := ResolveOAuthClient(t.Context(), clientID, client)
	if err != nil {
		t.Fatal(err)
	}
	if !c.AllowsRedirectURI("http://localhost:1234/callback", "") {
		t.Fatal("cimd loopback redirect denied")
	}
}

func TestValidateCIMDPinClaudeWeb(t *testing.T) {
	pin := pinnedCIMD["https://claude.ai/oauth/mcp-oauth-client-metadata"]
	if err := validateCIMD(pin.ClientID, &pin); err != nil {
		t.Fatal(err)
	}
	oc := oauthClient{Metadata: &pin, MetadataURL: pin.ClientID}
	if !oc.AllowsRedirectURI("https://claude.ai/api/mcp/auth_callback", "") {
		t.Fatal("claude web callback denied")
	}
}
