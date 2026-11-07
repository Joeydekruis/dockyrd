package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
	"github.com/Joeydekruis/dockyrd/internal/reconcile"
	"github.com/Joeydekruis/dockyrd/internal/serviceops"
)

type deleteServiceRequest struct {
	ConfirmSlug  string `json:"confirm_slug"`
	DeleteVolume bool   `json:"delete_volume"`
}

func registerServiceLifecycleRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("DELETE /api/v1/services/{id}", deps.handleDeleteService)
	mux.HandleFunc("POST /api/v1/services/{id}/restart", deps.handleRestartService)
}

func (d apiDeps) handleDeleteService(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	var body deleteServiceRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid json")
			return
		}
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
	if strings.TrimSpace(body.ConfirmSlug) != row.Slug {
		writeJSONError(w, http.StatusBadRequest, "confirm_slug must match the service slug")
		return
	}
	switch row.Kind {
	case "postgres", "redis":
		if err := serviceops.TeardownDatabase(ctx, d.docker, row.ID, row.Slug, body.DeleteVolume); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not remove database container")
			return
		}
	case "app":
		if err := serviceops.TeardownApplication(ctx, d.docker, row.ID, 30); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not remove application containers")
			return
		}
	case "static_site":
		if err := serviceops.TeardownStaticSite(d.cfg.SitesDir(), row.Slug); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "could not remove static site files")
			return
		}
	default:
		writeJSONError(w, http.StatusBadRequest, "this service kind cannot be deleted from the panel")
		return
	}
	if err := d.q.DeleteService(ctx, id); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "delete error")
		return
	}
	if err := proxy.SyncFromDB(ctx, d.q, d.cfg, d.docker); err != nil {
		// Best-effort; reconciler will reload Caddy on the next tick.
		_ = err
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "service.delete",
		TargetKind: "service",
		TargetID:   id,
		Meta: map[string]any{
			"kind":          row.Kind,
			"slug":          row.Slug,
			"delete_volume": body.DeleteVolume,
		},
	})
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleRestartService(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
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
	if row.Kind != "postgres" && row.Kind != "redis" {
		writeJSONError(w, http.StatusBadRequest, "restart applies to postgres and redis only")
		return
	}
	if row.DesiredState != "running" {
		writeJSONError(w, http.StatusBadRequest, "service is not running")
		return
	}
	if err := serviceops.StopManagedDatabaseContainer(ctx, d.docker, row.ID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "could not stop database container")
		return
	}
	reconcile.SyncDatabaseServices(ctx, d.cfg, d.q, d.docker, nil)
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "service.restart",
		TargetKind: "service",
		TargetID:   id,
		Meta:       map[string]any{"kind": row.Kind, "slug": row.Slug},
	})
	w.WriteHeader(http.StatusNoContent)
}
