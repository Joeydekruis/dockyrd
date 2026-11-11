package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func registerTeamRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/users", deps.handleListUsers)
	mux.HandleFunc("PATCH /api/v1/users/{id}", deps.handlePatchUser)
	mux.HandleFunc("GET /api/v1/invites", deps.handleListInvites)
	mux.HandleFunc("POST /api/v1/invites", deps.handleCreateInvite)
	mux.HandleFunc("DELETE /api/v1/invites/{id}", deps.handleDeleteInvite)
	mux.HandleFunc("GET /api/v1/audit", deps.handleListAudit)
}

type teamUserResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
	TotpEnabled  bool   `json:"totp_enabled"`
	DisabledAt   int64  `json:"disabled_at,omitempty"`
}

type inviteResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expires_at"`
	Token     string `json:"token,omitempty"`
	EmailSent bool   `json:"email_sent,omitempty"`
}

type auditResponse struct {
	ID         string         `json:"id"`
	UserID     string         `json:"user_id,omitempty"`
	Action     string         `json:"action"`
	TargetKind string         `json:"target_kind,omitempty"`
	TargetID   string         `json:"target_id,omitempty"`
	Meta       map[string]any `json:"meta,omitempty"`
	IP         string         `json:"ip,omitempty"`
	CreatedAt  int64          `json:"created_at"`
}

func (d apiDeps) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	rows, err := d.q.ListUsers(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]teamUserResponse, 0, len(rows))
	for _, row := range rows {
		u := teamUserResponse{
			ID:        row.ID,
			Email:     row.Email,
			Name:      row.Name,
			Role:      row.Role,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
			TotpEnabled: row.TotpEnabledAt.Valid,
		}
		if row.DisabledAt.Valid {
			u.DisabledAt = row.DisabledAt.Int64
		}
		out = append(out, u)
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": out})
}

func (d apiDeps) handlePatchUser(w http.ResponseWriter, r *http.Request) {
	actor, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if id == actor.ID {
		writeJSONError(w, http.StatusBadRequest, "cannot change your own role or remove yourself here")
		return
	}
	target, err := d.q.GetUserByID(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}
	if target.Role == "owner" {
		writeJSONError(w, http.StatusForbidden, "cannot modify the owner account")
		return
	}
	var body struct {
		Role     *string `json:"role"`
		Disabled *bool   `json:"disabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	ctx := r.Context()
	now := time.Now().Unix()
	if body.Disabled != nil && *body.Disabled {
		if err := d.q.DisableUser(ctx, sqlc.DisableUserParams{
			DisabledAt: sql.NullInt64{Int64: now, Valid: true},
			UpdatedAt:  now,
			ID:         id,
		}); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "database error")
			return
		}
		_ = d.q.DeleteAllSessionsForUser(ctx, id)
		_ = d.audit.Record(ctx, audit.Entry{
			UserID:     actor.ID,
			Action:     "user.remove",
			TargetKind: "user",
			TargetID:   id,
			Meta:       map[string]any{"email": target.Email},
			IP:         clientIP(r),
		})
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if body.Role == nil {
		writeJSONError(w, http.StatusBadRequest, "role or disabled required")
		return
	}
	role := strings.TrimSpace(*body.Role)
	if !validAssignRole(role) {
		writeJSONError(w, http.StatusBadRequest, "invalid role")
		return
	}
	if role == "owner" && actor.Role != "owner" {
		writeJSONError(w, http.StatusForbidden, "only owner can assign owner role")
		return
	}
	if err := d.q.UpdateUserRole(ctx, sqlc.UpdateUserRoleParams{
		Role:      role,
		UpdatedAt: now,
		ID:        id,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     actor.ID,
		Action:     "user.role_change",
		TargetKind: "user",
		TargetID:   id,
		Meta:       map[string]any{"role": role},
		IP:         clientIP(r),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleListInvites(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	rows, err := d.q.ListPendingInvites(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]inviteResponse, 0, len(rows))
	for _, row := range rows {
		inv := inviteResponse{ID: row.ID}
		if row.Email.Valid {
			inv.Email = row.Email.String
		}
		if row.Role.Valid {
			inv.Role = row.Role.String
		}
		if row.ExpiresAt.Valid {
			inv.ExpiresAt = row.ExpiresAt.Int64
		}
		out = append(out, inv)
	}
	writeJSON(w, http.StatusOK, map[string]any{"invites": out})
}

func (d apiDeps) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	actor, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	var body struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Role = strings.TrimSpace(body.Role)
	if !validAssignRole(body.Role) {
		writeJSONError(w, http.StatusBadRequest, "role required")
		return
	}
	if body.Role == "owner" && actor.Role != "owner" {
		writeJSONError(w, http.StatusForbidden, "only owner can invite as owner")
		return
	}
	plain, hash, err := auth.NewSessionToken()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "token error")
		return
	}
	id := ulid.Make().String()
	expires := time.Now().Add(7 * 24 * time.Hour).Unix()
	ctx := r.Context()
	emailParam := sql.NullString{}
	if body.Email != "" {
		emailParam = sql.NullString{String: body.Email, Valid: true}
	}
	if err := d.q.CreateInvite(ctx, sqlc.CreateInviteParams{
		ID:        id,
		Email:     emailParam,
		Role:      sql.NullString{String: body.Role, Valid: true},
		TokenHash: hash,
		InvitedBy: sql.NullString{String: actor.ID, Valid: true},
		ExpiresAt: sql.NullInt64{Int64: expires, Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     actor.ID,
		Action:     "invite.create",
		TargetKind: "invite",
		TargetID:   id,
		Meta:       map[string]any{"email": body.Email, "role": body.Role},
		IP:         clientIP(r),
	})
	emailSent := false
	smtpCfg, _, base, smtpOK := d.smtpReady(ctx)
	if smtpOK && body.Email != "" {
		link := "/accept-invite?token=" + plain
		if base != "" {
			link = base + link
		}
		subject := "dockyrd team invite"
		text := "You have been invited to join this dockyrd panel.\n\nAccept the invite:\n" + link + "\n"
		if err := d.sendPlainWithLog(ctx, emailKindTeamInvite, smtpCfg, body.Email, subject, text); err == nil {
			emailSent = true
		}
	}
	writeJSON(w, http.StatusCreated, inviteResponse{
		ID:        id,
		Email:     body.Email,
		Role:      body.Role,
		ExpiresAt: expires,
		Token:     plain,
		EmailSent: emailSent,
	})
}

func (d apiDeps) handleDeleteInvite(w http.ResponseWriter, r *http.Request) {
	actor, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	ctx := r.Context()
	if err := d.q.DeleteInvite(ctx, id); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     actor.ID,
		Action:     "invite.revoke",
		TargetKind: "invite",
		TargetID:   id,
		IP:         clientIP(r),
	})
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleListAudit(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	rows, err := d.q.ListAuditLogRecent(r.Context(), 100)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]auditResponse, 0, len(rows))
	for _, row := range rows {
		item := auditResponse{ID: row.ID}
		if row.UserID.Valid {
			item.UserID = row.UserID.String
		}
		if row.Action.Valid {
			item.Action = row.Action.String
		}
		if row.TargetKind.Valid {
			item.TargetKind = row.TargetKind.String
		}
		if row.TargetID.Valid {
			item.TargetID = row.TargetID.String
		}
		if row.Meta != nil {
			var meta map[string]any
			var raw []byte
			switch v := row.Meta.(type) {
			case []byte:
				raw = v
			case string:
				raw = []byte(v)
			}
			if len(raw) > 0 && json.Unmarshal(raw, &meta) == nil {
				item.Meta = meta
			}
		}
		if row.Ip.Valid {
			item.IP = row.Ip.String
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Int64
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": out})
}
