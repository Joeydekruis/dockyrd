package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

var allowedTerminalThemes = map[string]struct{}{
	"dockyrd-dark":   {},
	"dockyrd-light":  {},
	"hacker-green":   {},
	"normal-light":   {},
}

func terminalThemeSettingKey(userID string) string {
	return fmt.Sprintf("user_terminal_theme:%s", userID)
}

func (d apiDeps) loadTerminalTheme(ctx context.Context, userID string) string {
	const fallback = "dockyrd-dark"
	row, err := d.q.GetSetting(ctx, terminalThemeSettingKey(userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fallback
		}
		return fallback
	}
	raw, err := settingValueBytes(row.Value)
	if err != nil || len(raw) == 0 {
		return fallback
	}
	var body struct {
		ThemeID string `json:"theme_id"`
	}
	if json.Unmarshal(raw, &body) != nil {
		return fallback
	}
	id := strings.TrimSpace(body.ThemeID)
	if _, ok := allowedTerminalThemes[id]; !ok {
		return fallback
	}
	return id
}

func (d apiDeps) handleGetTerminalPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"theme_id": d.loadTerminalTheme(r.Context(), user.ID),
	})
}

func (d apiDeps) handlePatchTerminalPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	var body struct {
		ThemeID string `json:"theme_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	id := strings.TrimSpace(body.ThemeID)
	if _, ok := allowedTerminalThemes[id]; !ok {
		writeJSONError(w, http.StatusBadRequest, "invalid theme_id")
		return
	}
	raw, err := json.Marshal(map[string]string{"theme_id": id})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "encode error")
		return
	}
	if err := d.q.UpsertSetting(r.Context(), sqlc.UpsertSettingParams{
		Key:   terminalThemeSettingKey(user.ID),
		Value: string(raw),
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"theme_id": id})
}
