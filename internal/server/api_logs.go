package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/accesslog"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

func registerLogRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/logs", deps.handleServiceLogs)
	mux.HandleFunc("GET /api/v1/services/{id}/analytics", deps.handleServiceAnalytics)
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

	stream := strings.TrimSpace(r.URL.Query().Get("stream"))
	if stream == "" {
		if svc.Kind == "static_site" {
			stream = "access"
		} else {
			stream = "runtime"
		}
	}

	tail := r.URL.Query().Get("tail")
	if tail == "" {
		tail = "200"
	}
	tailN, _ := strconv.Atoi(tail)
	if tailN <= 0 {
		tailN = 200
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	switch stream {
	case "access":
		hosts, err := d.serviceHostnames(ctx, serviceID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "database error")
			return
		}
		text, err := accesslog.TailText(d.cfg.CaddyAccessLogPath(), accesslog.HostSet(hosts), tailN)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "log read error")
			return
		}
		_, _ = w.Write([]byte(text))
		return
	case "runtime":
		if svc.Kind == "static_site" {
			_, _ = w.Write([]byte("Static sites have no container runtime logs. Switch to access logs (stream=access)."))
			return
		}
	default:
		writeJSONError(w, http.StatusBadRequest, "stream must be runtime or access")
		return
	}

	if !svc.LiveDeploymentID.Valid {
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
		return
	}
	logs, err := cli.ContainerLogs(ctx, containerID, tail)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "log fetch error")
		return
	}
	_, _ = w.Write([]byte(logs))
}

func (d apiDeps) handleServiceAnalytics(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
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
	rangeQ := strings.TrimSpace(r.URL.Query().Get("range"))
	if rangeQ == "" {
		rangeQ = "24h"
	}
	since := metricRangeSince(rangeQ)
	hosts, err := d.serviceHostnames(ctx, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	sum, err := accesslog.AggregateSince(
		d.cfg.CaddyAccessLogPath(),
		accesslog.HostSet(hosts),
		time.Unix(since, 0).UTC(),
		rangeQ,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "analytics error")
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

func (d apiDeps) serviceHostnames(ctx context.Context, serviceID string) ([]string, error) {
	rows, err := d.q.ListDomainsForService(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		h := strings.TrimSpace(row.Hostname)
		if h != "" {
			out = append(out, h)
		}
	}
	return out, nil
}
