package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/project"
	"github.com/Joeydekruis/dockyrd/internal/storage"
)

func registerStorageRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/storage", deps.handleGetStorage)
	mux.HandleFunc("POST /api/v1/storage/install", deps.handleInstallStorage)
}

func (d apiDeps) handleGetStorage(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	installed, _ := storage.IsGarageInstalled(r.Context(), d.q)
	st, _ := storage.LoadGarage(r.Context(), d.q)
	buckets, err := d.q.ListServicesByKind(r.Context(), "bucket")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	bucketOut := make([]map[string]any, 0, len(buckets))
	for _, b := range buckets {
		cfg := serviceConfigJSON(b.Config)
		bucketOut = append(bucketOut, map[string]any{
			"id":           b.ID,
			"name":         b.Name,
			"slug":         b.Slug,
			"project_id":   b.ProjectID,
			"project_name": project.DisplayName(b.ProjectName, b.ProjectSlug),
			"project_slug": project.DisplaySlug(b.ProjectSlug),
			"config":       dbkind.RedactConfig("bucket", cfg),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"installed":         installed,
		"provider":          st.Provider,
		"version":           st.Version,
		"internal_endpoint": storage.InternalEndpoint,
		"buckets":           bucketOut,
	})
}

func (d apiDeps) handleInstallStorage(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	var body struct {
		Provider string `json:"provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Provider = strings.TrimSpace(strings.ToLower(body.Provider))
	if body.Provider == "" {
		body.Provider = "garage"
	}
	if err := storage.ValidateInstallRequest(body.Provider); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	cli, err := docker.NewClient()
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "docker unavailable")
		return
	}
	defer cli.Close()
	ctx := r.Context()
	if err := cli.Ping(ctx); err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "docker unavailable")
		return
	}
	if err := storage.InstallGarage(ctx, d.cfg, d.q, cli); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"installed":         true,
		"provider":          "garage",
		"internal_endpoint": storage.InternalEndpoint,
	})
}
