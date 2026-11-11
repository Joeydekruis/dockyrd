package oauth

import (
	"net/url"
	"strings"
)

// allowedRedirectURI permits loopback and registered native client schemes for MCP OAuth.
func allowedRedirectURI(redirectURI, panelOrigin string) bool {
	u, err := url.Parse(strings.TrimSpace(redirectURI))
	if err != nil || u.Scheme == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	switch u.Scheme {
	case "http", "https":
		if host == "127.0.0.1" || host == "localhost" {
			return true
		}
		if panelOrigin != "" {
			po, err := url.Parse(panelOrigin)
			if err == nil && strings.EqualFold(u.Host, po.Host) {
				return true
			}
		}
		return false
	case "cursor", "claude":
		return true
	default:
		return false
	}
}
