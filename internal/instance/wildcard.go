package instance

import (
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"strings"
)

// SuggestAppsWildcard returns *.apps.<apex> from the dashboard public URL host.
func SuggestAppsWildcard(publicURL string) string {
	host := PanelHostname(publicURL)
	if host == "" {
		return ""
	}
	apex := registrableDomain(host)
	if apex == "" {
		return ""
	}
	return "*." + appsWildcardSuffix(apex)
}

func appsWildcardSuffix(apex string) string {
	return "apps." + strings.ToLower(strings.TrimSpace(apex))
}

// WildcardSuffix strips a leading *. from a wildcard pattern.
func WildcardSuffix(pattern string) string {
	p := strings.TrimSpace(pattern)
	p = strings.TrimPrefix(p, "*.")
	return strings.TrimPrefix(p, ".")
}

// WildcardIsVerified returns true when verification timestamp is set.
func WildcardIsVerified(panel Panel) bool {
	return panel.WildcardVerifiedAt > 0 && strings.TrimSpace(panel.Wildcard) != ""
}

// ProbeHostnameForWildcard builds dockyrd-check.<suffix> for DNS verification.
func ProbeHostnameForWildcard(pattern string) string {
	suffix := WildcardSuffix(pattern)
	if suffix == "" {
		return ""
	}
	return "dockyrd-check." + suffix
}

// EnsureProbeHost sets WildcardProbeHost when empty or wildcard changed.
func EnsureProbeHost(panel *Panel) {
	if strings.TrimSpace(panel.Wildcard) == "" {
		panel.WildcardProbeHost = ""
		return
	}
	want := ProbeHostnameForWildcard(panel.Wildcard)
	if panel.WildcardProbeHost != want {
		panel.WildcardProbeHost = want
	}
}

func registrableDomain(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return host
	}
	if len(parts) == 2 {
		return host
	}
	return strings.Join(parts[len(parts)-2:], ".")
}

// NewWildcardProbeToken returns 4 hex chars for optional unique probes.
func NewWildcardProbeToken() (string, error) {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// PanelHostFromURL is PanelHostname exported for tests with scheme handling.
func PanelHostFromURL(raw string) string {
	u := strings.TrimSpace(raw)
	if u == "" {
		return ""
	}
	if !strings.Contains(u, "://") {
		u = "https://" + u
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(parsed.Hostname()))
}
