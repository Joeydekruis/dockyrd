package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func registerPasswordResetRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("POST /api/v1/auth/password/forgot", deps.handlePasswordForgot)
	mux.HandleFunc("POST /api/v1/auth/password/reset", deps.handlePasswordReset)
}

func (d apiDeps) handlePasswordForgot(w http.ResponseWriter, r *http.Request) {
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if body.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "email required")
		return
	}
	ctx := r.Context()
	user, err := d.q.GetUserByEmail(ctx, body.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	plain, hash, err := auth.NewSessionToken()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "token error")
		return
	}
	expires := time.Now().Add(time.Hour).Unix()
	if err := d.q.InsertPasswordReset(ctx, sqlc.InsertPasswordResetParams{
		TokenHash: hash,
		UserID:    sqlNullUserID(user.ID),
		ExpiresAt: sql.NullInt64{Int64: expires, Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	resp := map[string]any{"ok": true}
	smtpCfg, _, base, smtpOK := d.smtpReady(ctx)
	resp["smtp_configured"] = smtpOK
	if smtpOK {
		link := "/reset-password?token=" + plain
		if base != "" {
			link = base + link
		}
		text := "Use this link within one hour to choose a new password:\n\n" + link + "\n"
		subject := "Reset your Dockyard password"
		if err := d.sendPlainWithLog(ctx, emailKindPasswordReset, smtpCfg, body.Email, subject, text); err != nil {
			writeJSONError(w, http.StatusBadGateway, "could not send reset email")
			return
		}
		resp["email_sent"] = true
	}
	writeJSON(w, http.StatusOK, resp)
}

func (d apiDeps) handlePasswordReset(w http.ResponseWriter, r *http.Request) {
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Token = strings.TrimSpace(body.Token)
	if body.Token == "" || len(body.Password) < 12 {
		writeJSONError(w, http.StatusBadRequest, "token and password (12+ chars) required")
		return
	}
	ctx := r.Context()
	row, err := d.q.GetPasswordReset(ctx, auth.HashSessionToken(body.Token))
	if err != nil {
		writeJSONError(w, http.StatusForbidden, "invalid or expired reset link")
		return
	}
	if row.UsedAt.Valid {
		writeJSONError(w, http.StatusForbidden, "reset link already used")
		return
	}
	now := time.Now().Unix()
	if row.ExpiresAt.Valid && now > row.ExpiresAt.Int64 {
		writeJSONError(w, http.StatusForbidden, "invalid or expired reset link")
		return
	}
	if !row.UserID.Valid {
		writeJSONError(w, http.StatusForbidden, "invalid reset link")
		return
	}
	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "password hash error")
		return
	}
	if err := d.q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		PasswordHash: sqlNullString(hash),
		UpdatedAt:    now,
		ID:           row.UserID.String,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if err := d.q.MarkPasswordResetUsed(ctx, sqlc.MarkPasswordResetUsedParams{
		UsedAt:    sql.NullInt64{Int64: now, Valid: true},
		TokenHash: auth.HashSessionToken(body.Token),
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
