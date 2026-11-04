package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/slug"
	"github.com/Joeydekruis/dockyrd/internal/staticsite"
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

type patchServiceRequest struct {
	ProjectID string          `json:"project_id"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
}

func registerServiceRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services", deps.handleListServices)
	mux.HandleFunc("POST /api/v1/services", deps.handleCreateService)
	mux.HandleFunc("GET /api/v1/services/{id}", deps.handleGetService)
	mux.HandleFunc("PATCH /api/v1/services/{id}", deps.handlePatchService)
}

func (d apiDeps) handleListServices(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	projectID := strings.TrimSpace(r.URL.Query().Get("project_id"))
	if projectID == "" {
		writeJSONError(w, http.StatusBadRequest, "project_id query required")
		return
	}
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	rows, err := d.q.ListServicesByProject(r.Context(), projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]serviceResponse, 0, len(rows))
	for _, row := range rows {
		if kind != "" && row.Kind != kind {
			continue
		}
		out = append(out, serviceFromRow(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"services": out})
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
	if body.Kind != "app" && body.Kind != "static_site" && body.Kind != "postgres" && body.Kind != "redis" && body.Kind != "bucket" {
		writeJSONError(w, http.StatusBadRequest, "unsupported service kind")
		return
	}
	cfg := body.Config
	switch body.Kind {
	case "static_site":
		if len(cfg) == 0 {
			writeJSONError(w, http.StatusBadRequest, "config required for static_site")
			return
		}
	case "app":
		if len(cfg) == 0 {
			cfg = json.RawMessage(`{"source":{"type":"image","ref":"nginx:alpine"},"port":80}`)
		}
	case "postgres", "redis":
		masterKey, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "secrets not configured")
			return
		}
		version := parseConfigVersion(cfg)
		var genErr error
		if body.Kind == "postgres" {
			cfg, genErr = dbkind.NewPostgresConfigJSON(masterKey, version)
		} else {
			cfg, genErr = dbkind.NewRedisConfigJSON(masterKey, version)
		}
		if genErr != nil {
			writeJSONError(w, http.StatusInternalServerError, "config error")
			return
		}
	case "bucket":
		masterKey, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "secrets not configured")
			return
		}
		bucketName := slug.FromName(body.Name)
		publicRead := false
		if len(cfg) > 0 {
			var partial struct {
				BucketName string `json:"bucket_name"`
				PublicRead bool   `json:"public_read"`
			}
			_ = json.Unmarshal(cfg, &partial)
			if strings.TrimSpace(partial.BucketName) != "" {
				bucketName = partial.BucketName
			}
			publicRead = partial.PublicRead
		}
		cfg, err = dbkind.NewBucketConfigJSON(masterKey, bucketName, publicRead)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
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
	if body.Kind == "static_site" {
		d.provisionServiceDomains(ctx, row, cfg)
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

func (d apiDeps) handlePatchService(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	var body patchServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.ProjectID = strings.TrimSpace(body.ProjectID)
	body.Name = strings.TrimSpace(body.Name)
	hasProject := body.ProjectID != ""
	hasName := body.Name != ""
	hasConfig := len(body.Config) > 0 && string(body.Config) != "null"
	if !hasProject && !hasName && !hasConfig {
		writeJSONError(w, http.StatusBadRequest, "no fields to update")
		return
	}
	ctx := r.Context()
	row, err := d.q.GetServiceByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	if hasName || hasConfig {
		if row.Kind != "static_site" {
			writeJSONError(w, http.StatusBadRequest, "name and config updates apply to static_site only")
			return
		}
		now := time.Now().Unix()
		cfgJSON := serviceConfigJSON(row.Config)
		if hasConfig {
			merged, err := mergeServiceConfig(cfgJSON, body.Config)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "invalid config")
				return
			}
			if _, err := staticsite.Parse(merged); err != nil {
				writeJSONError(w, http.StatusBadRequest, err.Error())
				return
			}
			if err := d.q.UpdateServiceConfig(ctx, sqlc.UpdateServiceConfigParams{
				Config:    merged,
				UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
				ID:        id,
			}); err != nil {
				writeJSONError(w, http.StatusInternalServerError, "update error")
				return
			}
		}
		if hasName {
			if err := d.q.UpdateServiceName(ctx, sqlc.UpdateServiceNameParams{
				Name:      body.Name,
				UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
				ID:        id,
			}); err != nil {
				writeJSONError(w, http.StatusInternalServerError, "update error")
				return
			}
		}
	}

	if !hasProject {
		updated, err := d.q.GetServiceByID(ctx, id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "database error")
			return
		}
		writeJSON(w, http.StatusOK, serviceFromRow(updated))
		return
	}

	env, err := d.q.GetEnvironmentByProjectAndName(ctx, sqlc.GetEnvironmentByProjectAndNameParams{
		ProjectID: body.ProjectID,
		Name:      "production",
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "environment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if env.ID == row.EnvironmentID {
		updated, err := d.q.GetServiceByID(ctx, id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "database error")
			return
		}
		writeJSON(w, http.StatusOK, serviceFromRow(updated))
		return
	}
	if _, err := d.q.ServiceSlugExistsInEnvironment(ctx, sqlc.ServiceSlugExistsInEnvironmentParams{
		EnvironmentID: env.ID,
		Slug:          row.Slug,
	}); err == nil {
		writeJSONError(w, http.StatusConflict, "slug already used in target project")
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	now := time.Now().Unix()
	if err := d.q.UpdateServiceEnvironment(ctx, sqlc.UpdateServiceEnvironmentParams{
		EnvironmentID: env.ID,
		UpdatedAt:     sql.NullInt64{Int64: now, Valid: true},
		ID:            id,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update error")
		return
	}
	updated, err := d.q.GetServiceByID(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, serviceFromRow(updated))
}

func mergeServiceConfig(existing, patch json.RawMessage) (json.RawMessage, error) {
	var base map[string]json.RawMessage
	if err := json.Unmarshal(existing, &base); err != nil {
		return nil, err
	}
	if base == nil {
		base = map[string]json.RawMessage{}
	}
	var overlay map[string]json.RawMessage
	if err := json.Unmarshal(patch, &overlay); err != nil {
		return nil, err
	}
	for k, v := range overlay {
		base[k] = v
	}
	return json.Marshal(base)
}

func serviceFromRow(row sqlc.Service) serviceResponse {
	cfg := serviceConfigJSON(row.Config)
	display := dbkind.RedactConfig(row.Kind, cfg)
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
		Config:        display,
		CreatedAt:     created,
		UpdatedAt:     updated,
	}
}

func serviceConfigJSON(v interface{}) json.RawMessage {
	switch x := v.(type) {
	case json.RawMessage:
		return x
	case []byte:
		return json.RawMessage(x)
	case string:
		return json.RawMessage(x)
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return json.RawMessage(`{}`)
		}
		return b
	}
}

func parseConfigVersion(raw json.RawMessage) int {
	if len(raw) == 0 {
		return 0
	}
	var partial struct {
		Version int `json:"version"`
	}
	_ = json.Unmarshal(raw, &partial)
	return partial.Version
}
