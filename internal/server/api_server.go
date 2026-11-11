package server

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
	"github.com/Joeydekruis/dockyrd/internal/metrics"
)

func registerServerRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/server", deps.handleServerInfo)
	mux.HandleFunc("GET /api/v1/server/metrics", deps.handleServerMetrics)
	mux.HandleFunc("GET /api/v1/server/metrics/containers", deps.handleServerMetricsContainers)
	mux.HandleFunc("GET /api/v1/server/containers", deps.handleServerContainers)
	mux.HandleFunc("GET /api/v1/server/cleanup", deps.handleServerCleanupGet)
	mux.HandleFunc("POST /api/v1/server/cleanup", deps.handleServerCleanupPost)
	mux.HandleFunc("GET /api/v1/server/disk-usage", deps.handleServerDiskUsage)
	mux.HandleFunc("POST /api/v1/server/edge/reload", deps.handleServerEdgeReload)
}

type hostMetricPoint struct {
	TS        int64   `json:"ts"`
	CPU       float64 `json:"cpu"`
	MemUsed   int64   `json:"mem_used"`
	MemTotal  int64   `json:"mem_total"`
	DiskUsed  int64   `json:"disk_used"`
	DiskTotal int64   `json:"disk_total"`
	NetRx     int64   `json:"net_rx"`
	NetTx     int64   `json:"net_tx"`
	Load1     float64 `json:"load1"`
}

func (d apiDeps) handleServerMetrics(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	rangeQ := strings.TrimSpace(r.URL.Query().Get("range"))
	if rangeQ == "" {
		rangeQ = "24h"
	}
	since := time.Now().UTC().Add(-24 * time.Hour).Unix()
	switch rangeQ {
	case "30d":
		since = time.Now().UTC().Add(-30 * 24 * time.Hour).Unix()
	case "7d":
		since = time.Now().UTC().Add(-7 * 24 * time.Hour).Unix()
	case "1h":
		since = time.Now().UTC().Add(-time.Hour).Unix()
	}
	rows, err := d.q.ListMetricHostSince(r.Context(), sqlc.ListMetricHostSinceParams{
		Res: sql.NullInt64{Int64: eventlog.ResMinute, Valid: true},
		Ts:  sql.NullInt64{Int64: since, Valid: true},
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	points := make([]hostMetricPoint, 0, len(rows))
	for _, row := range rows {
		cpu := metrics.NormalizeHostCPUFraction(nullFloat64(row.Cpu))
		points = append(points, hostMetricPoint{
			TS:        nullInt64(row.Ts),
			CPU:       cpu,
			MemUsed:   nullInt64(row.MemUsed),
			MemTotal:  nullInt64(row.MemTotal),
			DiskUsed:  nullInt64(row.DiskUsed),
			DiskTotal: nullInt64(row.DiskTotal),
			NetRx:     nullInt64(row.NetRx),
			NetTx:     nullInt64(row.NetTx),
			Load1:     nullFloat64(row.Load1),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"range":   rangeQ,
		"points":  points,
		"res_sec": eventlog.ResMinute,
	})
}
