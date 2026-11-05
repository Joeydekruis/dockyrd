package server

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type projectLinkEdgeResponse struct {
	ID            string `json:"id"`
	FromServiceID string `json:"from_service_id"`
	ToServiceID   string `json:"to_service_id"`
	FromName      string `json:"from_name"`
	ToName        string `json:"to_name"`
	FromKind      string `json:"from_kind"`
	ToKind        string `json:"to_kind"`
	FromSlug      string `json:"from_slug"`
	ToSlug        string `json:"to_slug"`
}

func (d apiDeps) handleListProjectLinks(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	projectID := r.PathValue("id")
	ctx := r.Context()
	if _, err := d.q.GetProjectByID(ctx, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "project not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	rows, err := d.q.ListLinksForProject(ctx, projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	countRows, err := d.q.CountOutgoingLinksByProject(ctx, projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	counts := make(map[string]int64, len(countRows))
	for _, c := range countRows {
		counts[c.ServiceID] = c.LinkCount
	}
	cache := map[string]sqlc.Service{}
	lookup := func(id string) (sqlc.Service, error) {
		if s, ok := cache[id]; ok {
			return s, nil
		}
		s, err := d.q.GetServiceByID(ctx, id)
		if err != nil {
			return sqlc.Service{}, err
		}
		cache[id] = s
		return s, nil
	}
	edges := make([]projectLinkEdgeResponse, 0, len(rows))
	for _, row := range rows {
		from, err := lookup(row.FromServiceID)
		if err != nil {
			continue
		}
		to, err := lookup(row.ToServiceID)
		if err != nil {
			continue
		}
		edges = append(edges, projectLinkEdgeResponse{
			ID:            row.ID,
			FromServiceID: row.FromServiceID,
			ToServiceID:   row.ToServiceID,
			FromName:      from.Name,
			ToName:        to.Name,
			FromKind:      from.Kind,
			ToKind:        to.Kind,
			FromSlug:      from.Slug,
			ToSlug:        to.Slug,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"links":         edges,
		"outgoing_counts": counts,
	})
}
