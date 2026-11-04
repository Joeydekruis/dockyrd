package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/mail"
)

func registerInstanceSettingsRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/settings/public", deps.handleSettingsPublic)
	mux.HandleFunc("GET /api/v1/settings/instance", deps.handleGetInstanceSettings)
	mux.HandleFunc("PATCH /api/v1/settings/instance", deps.handlePatchInstanceSettings)
	mux.HandleFunc("POST /api/v1/settings/smtp/test", deps.handleSMTPTest)
}

func (d apiDeps) handleSettingsPublic(w http.ResponseWriter, r *http.Request) {
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	smtp, err := instance.LoadSMTPPublic(r.Context(), d.q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"smtp_configured": smtp.Configured,
	})
}

func (d apiDeps) handleGetInstanceSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	ctx := r.Context()
	panel, err := instance.LoadPanel(ctx, d.q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	smtp, err := instance.LoadSMTPPublic(ctx, d.q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"public_url":       panel.PublicURL,
		"acme_email":       panel.AcmeEmail,
		"wildcard_domain":  panel.Wildcard,
		"require_2fa_all":  panel.Require2FAAll,
		"smtp":             smtp,
	})
}

func (d apiDeps) handlePatchInstanceSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	var body struct {
		PublicURL       *string `json:"public_url"`
		AcmeEmail       *string `json:"acme_email"`
		WildcardDomain  *string `json:"wildcard_domain"`
		Require2FAAll   *bool   `json:"require_2fa_all"`
		SMTP            *struct {
			Host     string `json:"host"`
			Port     int    `json:"port"`
			Username string `json:"username"`
			From     string `json:"from"`
			UseTLS   bool   `json:"use_tls"`
			Password string `json:"password"`
		} `json:"smtp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	ctx := r.Context()
	panel, err := instance.LoadPanel(ctx, d.q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if body.PublicURL != nil {
		panel.PublicURL = strings.TrimSpace(*body.PublicURL)
	}
	if body.AcmeEmail != nil {
		panel.AcmeEmail = strings.TrimSpace(*body.AcmeEmail)
	}
	if body.WildcardDomain != nil {
		panel.Wildcard = strings.TrimSpace(*body.WildcardDomain)
	}
	if body.Require2FAAll != nil {
		panel.Require2FAAll = *body.Require2FAAll
	}
	if err := instance.SavePanel(ctx, d.q, panel); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	d.syncEdgeProxy(ctx)
	if body.SMTP != nil {
		pub := instance.SMTPPublic{
			Host:     strings.TrimSpace(body.SMTP.Host),
			Port:     body.SMTP.Port,
			Username: strings.TrimSpace(body.SMTP.Username),
			From:     strings.TrimSpace(body.SMTP.From),
			UseTLS:   body.SMTP.UseTLS,
		}
		if pub.Port == 0 {
			pub.Port = 587
		}
		if err := instance.SaveSMTP(ctx, d.q, d.cfg, pub, strings.TrimSpace(body.SMTP.Password)); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not save smtp settings")
			return
		}
	}
	d.handleGetInstanceSettings(w, r)
}

func (d apiDeps) handleSMTPTest(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	var body struct {
		To string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.To = strings.TrimSpace(strings.ToLower(body.To))
	if body.To == "" {
		writeJSONError(w, http.StatusBadRequest, "to email required")
		return
	}
	ctx := r.Context()
	cfg, ok, err := instance.LoadSMTPConfig(ctx, d.q, d.cfg)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "smtp not configured")
		return
	}
	subject := "Dockyard SMTP test"
	text := "If you received this message, SMTP is configured correctly."
	if err := mail.VerifyConnection(cfg); err != nil {
		d.recordEmailLog(ctx, emailKindSMTPTest, smtpFromAddress(cfg), body.To, subject, emailStatusFailed, err.Error(), cfg)
		writeJSONError(w, http.StatusBadGateway, "smtp host unreachable")
		return
	}
	if err := d.sendPlainWithLog(ctx, emailKindSMTPTest, cfg, body.To, subject, text); err != nil {
		writeJSONError(w, http.StatusBadGateway, "smtp send failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (d apiDeps) smtpReady(ctx context.Context) (instance.SMTPConfig, instance.Panel, string, bool) {
	panel, _ := instance.LoadPanel(ctx, d.q)
	base := instance.PublicBaseURL(d.cfg, panel)
	cfg, ok, _ := instance.LoadSMTPConfig(ctx, d.q, d.cfg)
	return cfg, panel, base, ok && cfg.Host != ""
}
