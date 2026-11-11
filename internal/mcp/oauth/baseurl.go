package oauth

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/instance"
)

// PanelOrigin returns the external origin for OAuth metadata (https://host).
func PanelOrigin(r *http.Request, cfg config.Config, panel instance.Panel) string {
	raw := strings.TrimSpace(panel.PublicURL)
	if raw == "" {
		raw = strings.TrimSpace(cfg.PublicURL())
	}
	if raw != "" {
		return alignOriginWithRequest(strings.TrimRight(raw, "/"), r)
	}
	return requestOrigin(r)
}

func requestOrigin(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := r.Host
	if h := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); h != "" {
		host = h
	}
	return scheme + "://" + strings.TrimRight(host, "/")
}

// alignOriginWithRequest upgrades http→https when the live request is HTTPS on the same host,
// so issuer/resource metadata match the URL users enter in Claude (common mis-set public URL).
func alignOriginWithRequest(configured string, r *http.Request) string {
	cfgURL, err := parseOrigin(configured)
	if err != nil {
		return configured
	}
	reqURL, err := parseOrigin(requestOrigin(r))
	if err != nil {
		return configured
	}
	if strings.EqualFold(cfgURL.Host, reqURL.Host) &&
		cfgURL.Scheme == "http" && reqURL.Scheme == "https" {
		cfgURL.Scheme = "https"
		return cfgURL.String()
	}
	return configured
}

func parseOrigin(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, err
	}
	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u, nil
}
