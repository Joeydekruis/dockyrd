package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func registerInvitePublicRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/invites/preview", deps.handleInvitePreview)
	mux.HandleFunc("POST /api/v1/invites/accept", deps.handleInviteAccept)
}

type invitePreviewResponse struct {
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
}

func (d apiDeps) handleInvitePreview(w http.ResponseWriter, r *http.Request) {
	inv, ok := d.inviteFromToken(w, r)
	if !ok {
		return
	}
	out := invitePreviewResponse{Role: "member"}
	if inv.Email.Valid {
		out.Email = inv.Email.String
	}
	if inv.Role.Valid {
		out.Role = inv.Role.String
	}
	if inv.ExpiresAt.Valid {
		out.ExpiresAt = inv.ExpiresAt.Int64
	}
	writeJSON(w, http.StatusOK, out)
}

func (d apiDeps) handleInviteAccept(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Token = strings.TrimSpace(body.Token)
	body.Name = strings.TrimSpace(body.Name)
	if body.Token == "" || body.Name == "" || len(body.Password) < 12 {
		writeJSONError(w, http.StatusBadRequest, "token, name, and password (12+ chars) required")
		return
	}

	ctx := r.Context()
	now := time.Now().Unix()
	hash := auth.HashSessionToken(body.Token)
	inv, err := d.q.GetPendingInviteByTokenHash(ctx, sqlc.GetPendingInviteByTokenHashParams{
		TokenHash: hash,
		ExpiresAt: sqlNullInt64(now),
	})
	if err != nil {
		writeJSONError(w, http.StatusForbidden, "invalid or expired invite")
		return
	}
	email := strings.ToLower(strings.TrimSpace(inv.Email.String))
	if email == "" {
		email = strings.ToLower(strings.TrimSpace(body.Email))
	}
	if email == "" {
		writeJSONError(w, http.StatusBadRequest, "email required for this invite")
		return
	}
	if _, err := d.q.GetUserByEmail(ctx, email); err == nil {
		writeJSONError(w, http.StatusConflict, "account already exists for this email")
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	role := "member"
	if inv.Role.Valid && validAssignRole(inv.Role.String) {
		role = inv.Role.String
	}
	passHash, err := auth.HashPassword(body.Password)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "password hash error")
		return
	}
	userID := ulid.Make().String()
	if err := d.q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           userID,
		Email:        email,
		Name:         body.Name,
		PasswordHash: sqlNullString(passHash),
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create user error")
		return
	}
	if err := d.q.MarkInviteAccepted(ctx, sqlc.MarkInviteAcceptedParams{
		AcceptedAt: sqlNullInt64(now),
		ID:         inv.ID,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "invite update error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     userID,
		Action:     "invite.accept",
		TargetKind: "invite",
		TargetID:   inv.ID,
		Meta:       map[string]any{"email": email, "role": role},
		IP:         clientIP(r),
	})
	if err := d.startSession(w, r, userID, true); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "session error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (d apiDeps) inviteFromToken(w http.ResponseWriter, r *http.Request) (sqlc.GetPendingInviteByTokenHashRow, bool) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		writeJSONError(w, http.StatusBadRequest, "token required")
		return sqlc.GetPendingInviteByTokenHashRow{}, false
	}
	now := time.Now().Unix()
	inv, err := d.q.GetPendingInviteByTokenHash(r.Context(), sqlc.GetPendingInviteByTokenHashParams{
		TokenHash: auth.HashSessionToken(token),
		ExpiresAt: sqlNullInt64(now),
	})
	if err != nil {
		writeJSONError(w, http.StatusForbidden, "invalid or expired invite")
		return sqlc.GetPendingInviteByTokenHashRow{}, false
	}
	return inv, true
}

func sqlNullInt64(n int64) sql.NullInt64 {
	return sql.NullInt64{Int64: n, Valid: true}
}

func sqlNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}
