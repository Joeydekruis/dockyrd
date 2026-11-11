package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const cimdFetchTimeout = 10 * time.Second

const cimdMaxBody = 64 << 10

// clientMetadata is a subset of OAuth Dynamic Client Registration metadata.
type clientMetadata struct {
	ClientID                    string   `json:"client_id"`
	ClientName                  string   `json:"client_name"`
	RedirectURIs                []string `json:"redirect_uris"`
	GrantTypes                  []string `json:"grant_types"`
	ResponseTypes               []string `json:"response_types"`
	TokenEndpointAuthMethod     string   `json:"token_endpoint_auth_method"`
}

// oauthClient is either a legacy opaque client_id or a resolved CIMD client.
type oauthClient struct {
	LegacyID   string
	Metadata   *clientMetadata
	MetadataURL string
}

func (c oauthClient) DisplayID() string {
	if c.MetadataURL != "" {
		return c.MetadataURL
	}
	return c.LegacyID
}

func (c oauthClient) AllowsRedirectURI(redirectURI, panelOrigin string) bool {
	if c.Metadata != nil {
		for _, allowed := range c.Metadata.RedirectURIs {
			if redirectURIMatches(allowed, redirectURI) && redirectURISafe(redirectURI) {
				return true
			}
		}
		return false
	}
	return allowedRedirectURI(redirectURI, panelOrigin)
}

var pinnedCIMD = map[string]clientMetadata{
	"https://claude.ai/oauth/mcp-oauth-client-metadata": {
		ClientID:                "https://claude.ai/oauth/mcp-oauth-client-metadata",
		ClientName:              "Claude",
		RedirectURIs:            []string{"https://claude.ai/api/mcp/auth_callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "none",
	},
	"https://claude.ai/api/oauth/mcp-oauth-client-metadata": {
		ClientID:                "https://claude.ai/api/oauth/mcp-oauth-client-metadata",
		ClientName:              "Claude",
		RedirectURIs:            []string{"https://claude.ai/api/mcp/auth_callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "none",
	},
	"https://claude.ai/oauth/claude-code-client-metadata": {
		ClientID:                "https://claude.ai/oauth/claude-code-client-metadata",
		ClientName:              "Claude Code",
		RedirectURIs:            []string{"http://localhost/callback", "http://127.0.0.1/callback"},
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "none",
	},
}

// ResolveOAuthClient loads CIMD when client_id is an https URL, otherwise treats it as legacy.
func ResolveOAuthClient(ctx context.Context, clientID string, httpClient *http.Client) (oauthClient, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return oauthClient{}, errors.New("empty client_id")
	}
	if !isCIMDClientID(clientID) {
		return oauthClient{LegacyID: clientID}, nil
	}
	meta, err := fetchCIMD(ctx, clientID, httpClient)
	if err != nil {
		return oauthClient{}, err
	}
	if err := validateCIMD(clientID, meta); err != nil {
		return oauthClient{}, err
	}
	return oauthClient{Metadata: meta, MetadataURL: clientID}, nil
}

func isCIMDClientID(clientID string) bool {
	u, err := url.Parse(clientID)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Scheme, "https") && u.Host != "" && u.Path != "" && u.Path != "/"
}

func fetchCIMD(ctx context.Context, clientID string, httpClient *http.Client) (*clientMetadata, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cimdFetchTimeout}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, clientID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, cimdMaxBody))
		if readErr != nil {
			return nil, readErr
		}
		var meta clientMetadata
		if jsonErr := json.Unmarshal(body, &meta); jsonErr != nil {
			return nil, jsonErr
		}
		return &meta, nil
	}
	if resp != nil {
		resp.Body.Close()
	}
	if pin, ok := pinnedCIMD[clientID]; ok {
		meta := pin
		return &meta, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("cimd fetch %s: status %d", clientID, resp.StatusCode)
}

func validateCIMD(clientID string, meta *clientMetadata) error {
	if meta == nil {
		return errors.New("missing metadata")
	}
	if meta.ClientID != clientID {
		return errors.New("client_id mismatch")
	}
	if len(meta.RedirectURIs) == 0 {
		return errors.New("missing redirect_uris")
	}
	if meta.TokenEndpointAuthMethod != "" && meta.TokenEndpointAuthMethod != "none" {
		return errors.New("unsupported token auth")
	}
	if len(meta.GrantTypes) > 0 && !sliceContains(meta.GrantTypes, "authorization_code") {
		return errors.New("grant_types")
	}
	if len(meta.ResponseTypes) > 0 && !sliceContains(meta.ResponseTypes, "code") {
		return errors.New("response_types")
	}
	return nil
}

func sliceContains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// redirectURIMatches implements loopback port-insensitive matching (RFC 8252).
func redirectURIMatches(registered, requested string) bool {
	if registered == requested {
		return true
	}
	reg, err1 := url.Parse(registered)
	req, err2 := url.Parse(requested)
	if err1 != nil || err2 != nil {
		return false
	}
	if !strings.EqualFold(reg.Scheme, req.Scheme) {
		return false
	}
	if reg.Path != req.Path {
		return false
	}
	if reg.RawQuery != req.RawQuery {
		return false
	}
	if isLoopbackHost(reg.Hostname()) && isLoopbackHost(req.Hostname()) {
		return strings.EqualFold(reg.Hostname(), req.Hostname())
	}
	return strings.EqualFold(reg.Host, req.Host)
}

func isLoopbackHost(host string) bool {
	switch strings.ToLower(host) {
	case "localhost", "127.0.0.1", "::1", "[::1]":
		return true
	default:
		return false
	}
}

// redirectURISafe enforces baseline redirect URI safety for CIMD clients.
func redirectURISafe(redirectURI string) bool {
	u, err := url.Parse(strings.TrimSpace(redirectURI))
	if err != nil || u.Scheme == "" {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		if isLoopbackHost(u.Hostname()) {
			return true
		}
		if u.Scheme == "https" && isKnownHostedCallbackHost(u.Hostname()) {
			return true
		}
		return false
	case "cursor", "claude":
		return true
	default:
		return false
	}
}

func isKnownHostedCallbackHost(host string) bool {
	switch strings.ToLower(host) {
	case "claude.ai", "claude.com":
		return true
	default:
		return false
	}
}
