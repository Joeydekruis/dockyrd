package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/dns"
	"github.com/Joeydekruis/dockyrd/internal/instance"
)

func (d apiDeps) instanceSettingsPayload(ctx context.Context, panel instance.Panel, smtp instance.SMTPPublic) map[string]any {
	instance.EnsureProbeHost(&panel)
	expected := d.cfg.EffectivePublicIPv4()
	wildcardStatus := "not_set"
	if strings.TrimSpace(panel.Wildcard) != "" {
		wildcardStatus = "unverified"
		if instance.WildcardIsVerified(panel) {
			wildcardStatus = "verified"
		} else if panel.WildcardProbeHost != "" && expected != "" {
			wildcardStatus = dns.EvaluateARecord(ctx, panel.WildcardProbeHost, expected)
			if wildcardStatus == "active" {
				wildcardStatus = "unverified"
			}
		}
	}
	return map[string]any{
		"public_url":                panel.PublicURL,
		"acme_email":                panel.AcmeEmail,
		"wildcard_domain":           panel.Wildcard,
		"wildcard_verified_at":      panel.WildcardVerifiedAt,
		"wildcard_verified":         instance.WildcardIsVerified(panel),
		"wildcard_dns_status":       wildcardStatus,
		"wildcard_probe_hostname":   panel.WildcardProbeHost,
		"suggested_apps_domain":     instance.SuggestAppsWildcard(panel.PublicURL),
		"public_ipv4":               expected,
		"require_2fa_all":           panel.Require2FAAll,
		"terminal_allow_host":       panel.TerminalAllowHost,
		"smtp":                      smtp,
	}
}

func registerInstanceWildcardRoute(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("POST /api/v1/settings/instance/verify-wildcard", deps.handleVerifyWildcard)
}

func (d apiDeps) handleVerifyWildcard(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	ctx := r.Context()
	panel, err := instance.LoadPanel(ctx, d.q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	pattern := strings.TrimSpace(panel.Wildcard)
	if pattern == "" {
		writeJSONError(w, http.StatusBadRequest, "apps domain not set")
		return
	}
	instance.EnsureProbeHost(&panel)
	probe := panel.WildcardProbeHost
	if probe == "" {
		writeJSONError(w, http.StatusBadRequest, "invalid apps domain pattern")
		return
	}
	expected := d.cfg.EffectivePublicIPv4()
	status := dns.EvaluateARecord(ctx, probe, expected)
	payload := map[string]any{
		"dns_status":              status,
		"wildcard_probe_hostname": probe,
		"expected_ipv4":           expected,
	}
	if status != "active" {
		payload["wildcard_verified"] = false
		payload["wildcard_verified_at"] = int64(0)
		writeJSON(w, http.StatusOK, payload)
		return
	}
	panel.WildcardVerifiedAt = time.Now().Unix()
	if err := instance.SavePanel(ctx, d.q, panel); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	payload["wildcard_verified"] = true
	payload["wildcard_verified_at"] = panel.WildcardVerifiedAt
	writeJSON(w, http.StatusOK, payload)
}

func applyInstancePanelPatch(panel *instance.Panel, prev instance.Panel, wildcardDomain *string, publicURL, acmeEmail *string, require2FA *bool) {
	if publicURL != nil {
		panel.PublicURL = strings.TrimSpace(*publicURL)
	}
	if acmeEmail != nil {
		panel.AcmeEmail = strings.TrimSpace(*acmeEmail)
	}
	if wildcardDomain != nil {
		next := strings.TrimSpace(*wildcardDomain)
		if !strings.EqualFold(strings.TrimSpace(prev.Wildcard), next) {
			panel.WildcardVerifiedAt = 0
		}
		panel.Wildcard = next
		instance.EnsureProbeHost(panel)
	}
	if require2FA != nil {
		panel.Require2FAAll = *require2FA
	}
}
