package server

import (
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

// allowedWebSocketOrigin rejects cross-site WebSocket hijacking when Origin is present.
func allowedWebSocketOrigin(r *http.Request, cfg config.Config) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}
	if pub := strings.TrimRight(strings.TrimSpace(cfg.PublicURL()), "/"); pub != "" {
		return strings.EqualFold(origin, pub)
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xf := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); xf != "" {
		scheme = strings.ToLower(strings.Split(xf, ",")[0])
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return false
	}
	expect := scheme + "://" + host
	return strings.EqualFold(origin, expect)
}
