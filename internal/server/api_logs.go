package server

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Joeydekruis/dockyrd/internal/docker"
)

func registerLogRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/logs", deps.handleServiceLogs)
}

func (d apiDeps) handleServiceLogs(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	svc, err := d.q.GetServiceByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if !svc.LiveDeploymentID.Valid {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(""))
		return
	}

	cli, err := docker.NewClient()
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "docker unavailable")
		return
	}
	defer cli.Close()

	containerID, err := cli.FindLiveContainerForService(ctx, serviceID, svc.LiveDeploymentID.String)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(""))
		return
	}
	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "200"
	}
	logs, err := cli.ContainerLogs(ctx, containerID, tail)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "log fetch error")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(logs))
}
