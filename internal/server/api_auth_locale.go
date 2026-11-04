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

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type localePrefs struct {
	Timezone   string `json:"timezone"`
	TimeFormat string `json:"time_format"`
}

func defaultLocalePrefs() localePrefs {
	return localePrefs{
		Timezone:   "UTC",
		TimeFormat: "24h",
	}
}

func localeSettingKey(userID string) string {
	return fmt.Sprintf("user_locale:%s", userID)
}

func registerAuthLocaleRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/auth/locale", deps.handleGetLocale)
	mux.HandleFunc("PATCH /api/v1/auth/locale", deps.handlePatchLocale)
}

func validateLocalePrefs(p localePrefs) error {
	if p.Timezone == "" {
		return fmt.Errorf("timezone required")
	}
	if _, err := time.LoadLocation(p.Timezone); err != nil {
		return fmt.Errorf("invalid timezone")
	}
	switch p.TimeFormat {
	case "24h", "12h":
	default:
		return fmt.Errorf("time_format must be 24h or 12h")
	}
	return nil
}

func (d apiDeps) loadLocalePrefs(ctx context.Context, userID string) (localePrefs, error) {
	prefs := defaultLocalePrefs()
	row, err := d.q.GetSetting(ctx, localeSettingKey(userID))
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
		return defaultLocalePrefs(), nil
	}
	if err := validateLocalePrefs(prefs); err != nil {
		return defaultLocalePrefs(), nil
	}
	return prefs, nil
}

func (d apiDeps) handleGetLocale(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	prefs, err := d.loadLocalePrefs(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

func (d apiDeps) handlePatchLocale(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	var body localePrefs
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Timezone = strings.TrimSpace(body.Timezone)
	if err := validateLocalePrefs(body); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	raw, err := json.Marshal(body)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "encode error")
		return
	}
	if err := d.q.UpsertSetting(r.Context(), sqlc.UpsertSettingParams{
		Key:   localeSettingKey(user.ID),
		Value: string(raw),
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, body)
}
