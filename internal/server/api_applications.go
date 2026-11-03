package server

import (
	"net/http"
)

func registerApplicationRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/applications", deps.handleListApplications)
}

type applicationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	DesiredState string `json:"desired_state"`
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	ProjectSlug string `json:"project_slug"`
}

func (d apiDeps) handleListApplications(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	rows, err := d.q.ListAppServices(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]applicationResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, applicationResponse{
			ID:           row.ID,
			Name:         row.Name,
			Slug:         row.Slug,
			DesiredState: row.DesiredState,
			ProjectID:    row.ProjectID,
			ProjectName:  row.ProjectName,
			ProjectSlug:  row.ProjectSlug,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"applications": out})
}
