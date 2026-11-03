package server

import (
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

type serviceResponse struct {
	ID            string          `json:"id"`
	EnvironmentID string          `json:"environment_id"`
	Kind          string          `json:"kind"`
	Name          string          `json:"name"`
	Slug          string          `json:"slug"`
	DesiredState  string          `json:"desired_state"`
	Config        json.RawMessage `json:"config"`
	CreatedAt     int64           `json:"created_at"`
	UpdatedAt     int64           `json:"updated_at"`
}

type createServiceRequest struct {
	ProjectID   string          `json:"project_id"`
	Environment string          `json:"environment"`
	Kind        string          `json:"kind"`
	Name        string          `json:"name"`
	Config      json.RawMessage `json:"config"`
}

func registerServiceRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("POST /api/v1/services", deps.handleCreateService)
	mux.HandleFunc("GET /api/v1/services/{id}", deps.handleGetService)
}

func (d apiDeps) handleCreateService(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	var body createServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.Kind = strings.TrimSpace(strings.ToLower(body.Kind))
	if body.Environment == "" {
		body.Environment = "production"
	}
	if body.ProjectID == "" || body.Name == "" || body.Kind == "" {
		writeJSONError(w, http.StatusBadRequest, "project_id, kind, and name required")
		return
	}
	if body.Kind != "app" {
		writeJSONError(w, http.StatusBadRequest, "only app services supported in this slice")
		return
	}
	cfg := body.Config
	if len(cfg) == 0 {
		cfg = json.RawMessage(`{"source":{"type":"image","ref":"nginx:alpine"},"port":80}`)
	}

	ctx := r.Context()
	env, err := d.q.GetEnvironmentByProjectAndName(ctx, sqlc.GetEnvironmentByProjectAndNameParams{
		ProjectID: body.ProjectID,
		Name:      body.Environment,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "environment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	now := time.Now().Unix()
	serviceID := ulid.Make().String()
	serviceSlug := slug.FromName(body.Name)
	if err := d.q.CreateService(ctx, sqlc.CreateServiceParams{
		ID:            serviceID,
		EnvironmentID: env.ID,
		Kind:          body.Kind,
		Name:          body.Name,
		Slug:          serviceSlug,
		DesiredState:  "running",
		Config:        cfg,
		CreatedAt:     sql.NullInt64{Int64: now, Valid: true},
		UpdatedAt:     sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create service error")
		return
	}
	row, err := d.q.GetServiceByID(ctx, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, serviceFromRow(row))
}

func (d apiDeps) handleGetService(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	row, err := d.q.GetServiceByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, serviceFromRow(row))
}

func serviceFromRow(row sqlc.Service) serviceResponse {
	cfg, _ := json.Marshal(row.Config)
	created := int64(0)
	if row.CreatedAt.Valid {
		created = row.CreatedAt.Int64
	}
	updated := int64(0)
	if row.UpdatedAt.Valid {
		updated = row.UpdatedAt.Int64
	}
	return serviceResponse{
		ID:            row.ID,
		EnvironmentID: row.EnvironmentID,
		Kind:          row.Kind,
		Name:          row.Name,
		Slug:          row.Slug,
		DesiredState:  row.DesiredState,
		Config:        cfg,
		CreatedAt:     created,
		UpdatedAt:     updated,
	}
}
