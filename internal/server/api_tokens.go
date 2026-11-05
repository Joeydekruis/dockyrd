package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/apitoken"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type apiTokenEntry struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Scope       string  `json:"scope"`
	ProjectID   string  `json:"project_id"`
	LastUsedAt  *int64  `json:"last_used_at,omitempty"`
	LastUsedIP  *string `json:"last_used_ip,omitempty"`
	CreatedAt   int64   `json:"created_at"`
	TokenPrefix string  `json:"token_prefix,omitempty"`
}

func registerTokenRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/tokens", deps.handleListAPITokens)
	mux.HandleFunc("POST /api/v1/tokens", deps.handleCreateAPIToken)
	mux.HandleFunc("DELETE /api/v1/tokens/{id}", deps.handleRevokeAPIToken)
}

func (d apiDeps) handleListAPITokens(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireSessionUser(w, r)
	if !ok {
		return
	}
	rows, err := d.q.ListAPITokensForUser(r.Context(), sql.NullString{String: user.ID, Valid: true})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]apiTokenEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, tokenEntryFromRow(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"tokens": out})
}

func (d apiDeps) handleCreateAPIToken(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireSessionUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Name      string `json:"name"`
		Scope     string `json:"scope"`
		ProjectID string `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.Scope = strings.TrimSpace(strings.ToLower(body.Scope))
	body.ProjectID = strings.TrimSpace(body.ProjectID)
	if body.Name == "" || body.ProjectID == "" {
		writeJSONError(w, http.StatusBadRequest, "name and project_id required")
		return
	}
	if !apitoken.ValidUserScope(body.Scope) {
		writeJSONError(w, http.StatusBadRequest, "scope must be read, deploy, or write")
		return
	}
	if _, err := d.q.GetProjectByID(r.Context(), body.ProjectID); err != nil {
		writeJSONError(w, http.StatusNotFound, "project not found")
		return
	}
	plain, hash, err := apitoken.Generate()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "token generation failed")
		return
	}
	now := time.Now().Unix()
	id := ulid.Make().String()
	if err := d.q.InsertAPIToken(r.Context(), sqlc.InsertAPITokenParams{
		ID:        id,
		Name:      sql.NullString{String: body.Name, Valid: true},
		TokenHash: hash,
		Scope:     body.Scope,
		CreatedBy: sql.NullString{String: user.ID, Valid: true},
		CreatedAt: now,
		ProjectID: sql.NullString{String: body.ProjectID, Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "store token failed")
		return
	}
	entry := apiTokenEntry{
		ID:          id,
		Name:        body.Name,
		Scope:       body.Scope,
		ProjectID:   body.ProjectID,
		CreatedAt:   now,
		TokenPrefix: apitoken.Prefix(plain),
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"token": plain,
		"entry": entry,
	})
}

func (d apiDeps) handleRevokeAPIToken(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireSessionUser(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	now := time.Now().Unix()
	if err := d.q.RevokeAPIToken(r.Context(), sqlc.RevokeAPITokenParams{
		RevokedAt: sql.NullInt64{Int64: now, Valid: true},
		ID:        id,
		CreatedBy: sql.NullString{String: user.ID, Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "revoke failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func tokenEntryFromRow(row sqlc.ListAPITokensForUserRow) apiTokenEntry {
	entry := apiTokenEntry{
		ID:        row.ID,
		Scope:     row.Scope,
		CreatedAt: row.CreatedAt,
	}
	if row.Name.Valid {
		entry.Name = row.Name.String
	}
	if row.ProjectID.Valid {
		entry.ProjectID = row.ProjectID.String
	}
	if row.LastUsedAt.Valid {
		v := row.LastUsedAt.Int64
		entry.LastUsedAt = &v
	}
	if row.LastUsedIp.Valid {
		s := row.LastUsedIp.String
		entry.LastUsedIP = &s
	}
	if len(row.TokenHash) >= 4 {
		entry.TokenPrefix = "dyk_…"
	}
	return entry
}
