package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/mail"
)

const (
	emailKindSMTPTest       = "smtp_test"
	emailKindPasswordReset  = "password_reset"
	emailKindTeamInvite     = "team_invite"
	emailStatusSent         = "sent"
	emailStatusFailed       = "failed"
	emailLogDefaultLimit    = 100
	emailLogMaxLimit        = 200
)

func registerEmailLogRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/settings/email/log", deps.handleListEmailLog)
	mux.HandleFunc("GET /api/v1/settings/email/log/{id}", deps.handleGetEmailLog)
}

type emailLogResponse struct {
	ID           string         `json:"id"`
	Kind         string         `json:"kind"`
	From         string         `json:"from"`
	To           string         `json:"to"`
	Subject      string         `json:"subject"`
	Status       string         `json:"status"`
	Response     string         `json:"response"`
	Detail       map[string]any `json:"detail,omitempty"`
	CreatedAt    int64          `json:"created_at"`
	CreatedAtISO string         `json:"created_at_iso"`
}

func (d apiDeps) recordEmailLog(ctx context.Context, kind, from, to, subject, status, response string, cfg instance.SMTPConfig) {
	if d.q == nil {
		return
	}
	detailBytes, _ := json.Marshal(map[string]any{
		"smtp_host": cfg.Host,
		"smtp_port": cfg.Port,
		"use_tls":   cfg.UseTLS,
	})
	_ = d.q.InsertEmailLog(ctx, sqlc.InsertEmailLogParams{
		ID:           ulid.Make().String(),
		Kind:         kind,
		FromAddress:  sql.NullString{String: from, Valid: from != ""},
		ToAddress:    to,
		Subject:      sql.NullString{String: subject, Valid: subject != ""},
		Status:       status,
		ResponseText: sql.NullString{String: response, Valid: response != ""},
		DetailLog:    sql.NullString{String: string(detailBytes), Valid: len(detailBytes) > 0},
		CreatedAt:    time.Now().Unix(),
	})
}

func smtpFromAddress(cfg instance.SMTPConfig) string {
	if cfg.From != "" {
		return cfg.From
	}
	return cfg.Username
}

func (d apiDeps) sendPlainWithLog(ctx context.Context, kind string, cfg instance.SMTPConfig, to, subject, body string) error {
	from := smtpFromAddress(cfg)
	err := mail.SendPlain(cfg, to, subject, body)
	status := emailStatusSent
	response := "Message accepted for delivery"
	if err != nil {
		status = emailStatusFailed
		response = err.Error()
	}
	d.recordEmailLog(ctx, kind, from, to, subject, status, response, cfg)
	return err
}

func emailLogRowToResponse(row sqlc.EmailLog) emailLogResponse {
	out := emailLogResponse{
		ID:        row.ID,
		Kind:      row.Kind,
		To:        row.ToAddress,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
	}
	if row.FromAddress.Valid {
		out.From = row.FromAddress.String
	}
	if row.Subject.Valid {
		out.Subject = row.Subject.String
	}
	if row.ResponseText.Valid {
		out.Response = row.ResponseText.String
	}
	if row.DetailLog.Valid && row.DetailLog.String != "" {
		var detail map[string]any
		if err := json.Unmarshal([]byte(row.DetailLog.String), &detail); err == nil {
			out.Detail = detail
		}
	}
	out.CreatedAtISO = time.Unix(row.CreatedAt, 0).UTC().Format(time.RFC3339)
	return out
}

func (d apiDeps) handleListEmailLog(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	limit := int64(emailLogDefaultLimit)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = int64(n)
		}
	}
	if limit > emailLogMaxLimit {
		limit = emailLogMaxLimit
	}
	rows, err := d.q.ListEmailLogRecent(r.Context(), limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]emailLogResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, emailLogRowToResponse(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": out})
}

func (d apiDeps) handleGetEmailLog(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id required")
		return
	}
	row, err := d.q.GetEmailLog(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, emailLogRowToResponse(row))
}
