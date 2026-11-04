package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type notificationPrefs struct {
	Deployments    bool `json:"deployments"`
	SecurityAlerts bool `json:"security_alerts"`
	ProductUpdates bool `json:"product_updates"`
}

func defaultNotificationPrefs() notificationPrefs {
	return notificationPrefs{
		Deployments:    true,
		SecurityAlerts: true,
		ProductUpdates: false,
	}
}

func notificationSettingKey(userID string) string {
	return fmt.Sprintf("user_notifications:%s", userID)
}

func registerAuthAccountRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("PATCH /api/v1/auth/account", deps.handlePatchAccount)
	mux.HandleFunc("POST /api/v1/auth/password/change", deps.handlePasswordChange)
	mux.HandleFunc("GET /api/v1/auth/notifications", deps.handleGetNotifications)
	mux.HandleFunc("PATCH /api/v1/auth/notifications", deps.handlePatchNotifications)
	registerAuthLocaleRoutes(mux, deps)
}

func (d apiDeps) handlePatchAccount(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Name  *string `json:"name"`
		Email *string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	name := user.Name
	email := user.Email
	if body.Name != nil {
		name = strings.TrimSpace(*body.Name)
		if name == "" {
			writeJSONError(w, http.StatusBadRequest, "name required")
			return
		}
	}
	if body.Email != nil {
		email = strings.TrimSpace(strings.ToLower(*body.Email))
		if email == "" {
			writeJSONError(w, http.StatusBadRequest, "email required")
			return
		}
	}
	ctx := r.Context()
	now := time.Now().Unix()
	if err := d.q.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		Name:      name,
		Email:     email,
		UpdatedAt: now,
		ID:        user.ID,
	}); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			writeJSONError(w, http.StatusConflict, "email already in use")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "user.profile_update",
		TargetKind: "user",
		TargetID:   user.ID,
		IP:         clientIP(r),
	})
	writeJSON(w, http.StatusOK, userResponse{
		ID:          user.ID,
		Email:       email,
		Name:        name,
		Role:        user.Role,
		TOTPEnabled: totpEnabled(user.TotpEnabledAt),
	})
}

func (d apiDeps) handlePasswordChange(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(body.NewPassword) < 12 {
		writeJSONError(w, http.StatusBadRequest, "new password must be at least 12 characters")
		return
	}
	ctx := r.Context()
	row, err := d.q.GetUserByEmail(ctx, user.Email)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if !row.PasswordHash.Valid {
		writeJSONError(w, http.StatusBadRequest, "password not set")
		return
	}
	match, err := auth.VerifyPassword(body.CurrentPassword, row.PasswordHash.String)
	if err != nil || !match {
		writeJSONError(w, http.StatusUnauthorized, "current password is incorrect")
		return
	}
	hash, err := auth.HashPassword(body.NewPassword)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "password hash error")
		return
	}
	now := time.Now().Unix()
	if err := d.q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		PasswordHash: sql.NullString{String: hash, Valid: true},
		UpdatedAt:    now,
		ID:           user.ID,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "user.password_change",
		TargetKind: "user",
		TargetID:   user.ID,
		IP:         clientIP(r),
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (d apiDeps) loadNotificationPrefs(ctx context.Context, userID string) (notificationPrefs, error) {
	prefs := defaultNotificationPrefs()
	row, err := d.q.GetSetting(ctx, notificationSettingKey(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return prefs, nil
		}
		return prefs, err
	}
	raw, err := settingValueBytes(row.Value)
	if err != nil {
		return prefs, err
	}
	if len(raw) == 0 {
		return prefs, nil
	}
	if err := json.Unmarshal(raw, &prefs); err != nil {
		return defaultNotificationPrefs(), nil
	}
	return prefs, nil
}

func (d apiDeps) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	prefs, err := d.loadNotificationPrefs(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

func (d apiDeps) handlePatchNotifications(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	var body notificationPrefs
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	raw, err := json.Marshal(body)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "encode error")
		return
	}
	if err := d.q.UpsertSetting(r.Context(), sqlc.UpsertSettingParams{
		Key:   notificationSettingKey(user.ID),
		Value: string(raw),
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, body)
}
