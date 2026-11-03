package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/slug"
)

type projectResponse struct {
	ID          string  `json:"id"`
	ServerID    string  `json:"server_id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`
}

type createProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type patchProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func registerProjectRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/projects", deps.handleListProjects)
	mux.HandleFunc("POST /api/v1/projects", deps.handleCreateProject)
	mux.HandleFunc("GET /api/v1/projects/{id}", deps.handleGetProject)
	mux.HandleFunc("PATCH /api/v1/projects/{id}", deps.handlePatchProject)
	mux.HandleFunc("DELETE /api/v1/projects/{id}", deps.handleDeleteProject)
}

func (d apiDeps) handleListProjects(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	rows, err := d.q.ListProjects(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]projectResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, projectFromRow(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"projects": out})
}

func (d apiDeps) handleGetProject(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	row, err := d.q.GetProjectByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "project not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, projectFromRow(row))
}

func (d apiDeps) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	var body createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name required")
		return
	}
	if len(body.Name) > 100 {
		writeJSONError(w, http.StatusBadRequest, "name too long")
		return
	}

	ctx := r.Context()
	srv, err := EnsureLocalServer(ctx, d.q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "server bootstrap error")
		return
	}
	projectSlug, err := d.uniqueProjectSlug(ctx, slug.FromName(body.Name))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "slug error")
		return
	}

	now := time.Now().Unix()
	projectID := ulid.Make().String()
	desc := sql.NullString{}
	if body.Description != nil {
		desc = sql.NullString{String: strings.TrimSpace(*body.Description), Valid: true}
	}
	if err := d.q.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:          projectID,
		ServerID:    srv.ID,
		Name:        body.Name,
		Slug:        projectSlug,
		Description: desc,
		CreatedAt:   sql.NullInt64{Int64: now, Valid: true},
		UpdatedAt:   sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create project error")
		return
	}
	envID := ulid.Make().String()
	if err := d.q.CreateEnvironment(ctx, sqlc.CreateEnvironmentParams{
		ID:        envID,
		ProjectID: projectID,
		Name:      "production",
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create environment error")
		return
	}
	if err := d.q.AddProjectMember(ctx, sqlc.AddProjectMemberParams{
		ProjectID: sql.NullString{String: projectID, Valid: true},
		UserID:    sql.NullString{String: user.ID, Valid: true},
		Role:      sql.NullString{String: "admin", Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "project member error")
		return
	}

	row, err := d.q.GetProjectByID(ctx, projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, projectFromRow(row))
}

func (d apiDeps) handlePatchProject(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	var body patchProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	ctx := r.Context()
	row, err := d.q.GetProjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "project not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	name := row.Name
	if body.Name != nil {
		name = strings.TrimSpace(*body.Name)
		if name == "" {
			writeJSONError(w, http.StatusBadRequest, "name required")
			return
		}
	}
	desc := row.Description
	if body.Description != nil {
		desc = sql.NullString{String: strings.TrimSpace(*body.Description), Valid: true}
	}
	now := time.Now().Unix()
	if err := d.q.UpdateProject(ctx, sqlc.UpdateProjectParams{
		Name:        name,
		Description: desc,
		UpdatedAt:   sql.NullInt64{Int64: now, Valid: true},
		ID:          id,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update error")
		return
	}
	updated, err := d.q.GetProjectByID(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, projectFromRow(updated))
}

func (d apiDeps) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	if err := d.q.DeleteProject(r.Context(), id); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "delete error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) requireUser(w http.ResponseWriter, r *http.Request) (sqlc.GetUserByIDRow, bool) {
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return sqlc.GetUserByIDRow{}, false
	}
	user, err := d.userFromRequest(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return sqlc.GetUserByIDRow{}, false
	}
	return user, true
}

func (d apiDeps) uniqueProjectSlug(ctx context.Context, base string) (string, error) {
	candidate := base
	for i := 0; i < 100; i++ {
		_, err := d.q.GetProjectBySlug(ctx, candidate)
		if errors.Is(err, sql.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
		candidate = base + "-" + ulid.Make().String()[22:]
	}
	return "", errors.New("could not allocate slug")
}

func projectFromRow(row sqlc.Project) projectResponse {
	var desc *string
	if row.Description.Valid && row.Description.String != "" {
		s := row.Description.String
		desc = &s
	}
	created := int64(0)
	if row.CreatedAt.Valid {
		created = row.CreatedAt.Int64
	}
	updated := int64(0)
	if row.UpdatedAt.Valid {
		updated = row.UpdatedAt.Int64
	}
	return projectResponse{
		ID:          row.ID,
		ServerID:    row.ServerID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: desc,
		CreatedAt:   created,
		UpdatedAt:   updated,
	}
}
