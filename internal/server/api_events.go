package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func registerEventRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/events", deps.handleListEvents)
	mux.HandleFunc("POST /api/v1/events/{id}/ack", deps.handleAckEvent)
}

type eventResponse struct {
	ID             string `json:"id"`
	Level          string `json:"level"`
	Category       string `json:"category"`
	Title          string `json:"title"`
	Detail         string `json:"detail,omitempty"`
	ProjectID      string `json:"project_id,omitempty"`
	ServiceID      string `json:"service_id,omitempty"`
	DeploymentID   string `json:"deployment_id,omitempty"`
	AcknowledgedAt int64  `json:"acknowledged_at,omitempty"`
	CreatedAt      int64  `json:"created_at"`
}

func (d apiDeps) handleListEvents(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	limit := int64(50)
	if v := strings.TrimSpace(r.URL.Query().Get("limit")); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	rows, err := d.q.ListRecentEvents(r.Context(), limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]eventResponse, 0, len(rows))
	for _, row := range rows {
		ev := eventFromRow(row)
		if category != "" && category != "all" && ev.Category != category {
			continue
		}
		out = append(out, ev)
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": out})
}

func (d apiDeps) handleAckEvent(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	ctx := r.Context()
	if _, err := d.q.GetEventByID(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "event not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	now := time.Now().Unix()
	if err := d.q.AckEvent(ctx, sqlc.AckEventParams{
		AcknowledgedAt: sql.NullInt64{Int64: now, Valid: true},
		ID:             id,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func eventFromRow(row sqlc.Event) eventResponse {
	ev := eventResponse{ID: row.ID}
	if row.Level.Valid {
		ev.Level = row.Level.String
	}
	if row.Category.Valid {
		ev.Category = row.Category.String
	}
	if row.Title.Valid {
		ev.Title = row.Title.String
	}
	if row.Detail.Valid {
		ev.Detail = row.Detail.String
	}
	if row.ProjectID.Valid {
		ev.ProjectID = row.ProjectID.String
	}
	if row.ServiceID.Valid {
		ev.ServiceID = row.ServiceID.String
	}
	if row.DeploymentID.Valid {
		ev.DeploymentID = row.DeploymentID.String
	}
	if row.AcknowledgedAt.Valid {
		ev.AcknowledgedAt = row.AcknowledgedAt.Int64
	}
	if row.CreatedAt.Valid {
		ev.CreatedAt = row.CreatedAt.Int64
	}
	return ev
}
