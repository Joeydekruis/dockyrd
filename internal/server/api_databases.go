package server

import (
	"encoding/json"
	"net/http"

	"github.com/Joeydekruis/dockyrd/internal/dbkind"
)

func registerDatabaseRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/databases", deps.handleListDatabases)
	registerDatabaseQueryRoutes(mux, deps)
}

type databaseRowResponse struct {
	ID           string          `json:"id"`
	Kind         string          `json:"kind"`
	Name         string          `json:"name"`
	Slug         string          `json:"slug"`
	DesiredState string          `json:"desired_state"`
	Config       json.RawMessage `json:"config"`
	ProjectID    string          `json:"project_id"`
	ProjectName  string          `json:"project_name"`
	ProjectSlug  string          `json:"project_slug"`
}

func (d apiDeps) handleListDatabases(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	rows, err := d.q.ListDatabaseServices(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]databaseRowResponse, 0, len(rows))
	for _, row := range rows {
		cfg := serviceConfigJSON(row.Config)
		out = append(out, databaseRowResponse{
			ID:           row.ID,
			Kind:         row.Kind,
			Name:         row.Name,
			Slug:         row.Slug,
			DesiredState: row.DesiredState,
			Config:       dbkind.RedactConfig(row.Kind, cfg),
			ProjectID:    row.ProjectID,
			ProjectName:  row.ProjectName,
			ProjectSlug:  row.ProjectSlug,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"databases": out})
}
