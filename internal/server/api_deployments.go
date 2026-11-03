package server

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
)

type deploymentResponse struct {
	ID         string  `json:"id"`
	ServiceID  string  `json:"service_id"`
	Status     string  `json:"status"`
	Trigger    string  `json:"trigger"`
	ImageRef   *string `json:"image_ref,omitempty"`
	Error      *string `json:"error,omitempty"`
	QueuedAt   int64   `json:"queued_at"`
	StartedAt  *int64  `json:"started_at,omitempty"`
	FinishedAt *int64  `json:"finished_at,omitempty"`
}

func registerDeploymentRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/deployments", deps.handleListDeployments)
	mux.HandleFunc("POST /api/v1/services/{id}/deployments", deps.handleCreateDeployment)
	mux.HandleFunc("POST /api/v1/deployments/{id}/cancel", deps.handleCancelDeployment)
}

func (d apiDeps) handleListDeployments(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	if _, err := d.q.GetServiceByID(r.Context(), serviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	rows, err := d.q.ListDeploymentsForService(r.Context(), serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]deploymentResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, deploymentFromRow(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"deployments": out})
}

func (d apiDeps) handleCreateDeployment(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	if _, err := d.q.GetServiceByID(ctx, serviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	now := time.Now().Unix()
	deploymentID := ulid.Make().String()
	if err := d.q.CreateDeployment(ctx, sqlc.CreateDeploymentParams{
		ID:          deploymentID,
		ServiceID:   serviceID,
		Status:      "queued",
		Trigger:     "manual",
		TriggeredBy: sql.NullString{String: user.ID, Valid: true},
		ImageRef:    sql.NullString{},
		QueuedAt:    sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create deployment error")
		return
	}
	if err := jobs.EnqueueDeploy(ctx, d.q, deploymentID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "enqueue job error")
		return
	}
	if d.events != nil {
		d.events.Publish("deployments:"+deploymentID, map[string]string{"status": "queued"})
	}
	row, err := d.q.GetDeploymentByID(ctx, deploymentID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, deploymentFromFullRow(row))
}

func (d apiDeps) handleCancelDeployment(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	ctx := r.Context()
	row, err := d.q.GetDeploymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if row.Status != "queued" {
		writeJSONError(w, http.StatusConflict, "deployment not cancellable")
		return
	}
	finished := time.Now().Unix()
	if err := d.q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status:    "cancelled",
		Error:     sql.NullString{},
		StartedAt: sql.NullInt64{},
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ID:        id,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func deploymentFromRow(row sqlc.ListDeploymentsForServiceRow) deploymentResponse {
	return deploymentResponse{
		ID:         row.ID,
		ServiceID:  row.ServiceID,
		Status:     row.Status,
		Trigger:    row.Trigger,
		ImageRef:   nullStrPtr(row.ImageRef),
		Error:      nullStrPtr(row.Error),
		QueuedAt:   nullInt64(row.QueuedAt),
		StartedAt:  nullInt64Ptr(row.StartedAt),
		FinishedAt: nullInt64Ptr(row.FinishedAt),
	}
}

func deploymentFromFullRow(row sqlc.Deployment) deploymentResponse {
	return deploymentResponse{
		ID:         row.ID,
		ServiceID:  row.ServiceID,
		Status:     row.Status,
		Trigger:    row.Trigger,
		ImageRef:   nullStrPtr(row.ImageRef),
		Error:      nullStrPtr(row.Error),
		QueuedAt:   nullInt64(row.QueuedAt),
		StartedAt:  nullInt64Ptr(row.StartedAt),
		FinishedAt: nullInt64Ptr(row.FinishedAt),
	}
}

func nullStrPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func nullInt64(v sql.NullInt64) int64 {
	if !v.Valid {
		return 0
	}
	return v.Int64
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	n := v.Int64
	return &n
}
