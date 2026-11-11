package oauth

import (
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/instance"
)

// PanelOrigin returns the external origin for OAuth metadata (https://host).
func PanelOrigin(r *http.Request, cfg config.Config, panel instance.Panel) string {
	if u := strings.TrimSpace(panel.PublicURL); u != "" {
		return strings.TrimRight(u, "/")
	}
	if u := strings.TrimSpace(cfg.PublicURL()); u != "" {
		return strings.TrimRight(u, "/")
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := r.Host
	if h := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); h != "" {
		host = h
	}
	return scheme + "://" + host
}
