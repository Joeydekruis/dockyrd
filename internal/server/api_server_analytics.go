package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
)

type containerAPIRow struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	State         string  `json:"state"`
	Kind          string  `json:"kind,omitempty"`
	ServiceID     string  `json:"service_id,omitempty"`
	CreatedAt     int64   `json:"created_at,omitempty"`
	ProjectLabel  string  `json:"project_label,omitempty"`
	CPUPercent    float64 `json:"cpu_percent,omitempty"`
	MemoryBytes   int64   `json:"memory_bytes,omitempty"`
	MemoryLimit   int64   `json:"memory_limit,omitempty"`
}

type containerMetricRow struct {
	ServiceID    string  `json:"service_id"`
	Name         string  `json:"name,omitempty"`
	ProjectLabel string  `json:"project_label,omitempty"`
	TS           int64   `json:"ts"`
	CPUPercent   float64 `json:"cpu_percent"`
	MemoryBytes  int64   `json:"memory_bytes"`
	MemoryLimit  int64   `json:"memory_limit"`
	NetRx        int64   `json:"net_rx"`
	NetTx        int64   `json:"net_tx"`
}

func (d apiDeps) handleServerMetricsContainers(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	rangeQ := strings.TrimSpace(r.URL.Query().Get("range"))
	if rangeQ == "" {
		rangeQ = "24h"
	}
	since := metricRangeSince(rangeQ)
	latest, err := d.q.ListLatestContainerMetrics(r.Context(), sqlc.ListLatestContainerMetricsParams{
		Res:   sql.NullInt64{Int64: eventlog.ResMinute, Valid: true},
		Res_2: sql.NullInt64{Int64: eventlog.ResMinute, Valid: true},
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	names := d.containerNameByService(r.Context())
	out := make([]containerMetricRow, 0, len(latest))
	for _, row := range latest {
		if !row.ServiceID.Valid {
			continue
		}
		sid := row.ServiceID.String
		cpu := nullFloat64(row.Cpu)
		if cpu > 1 {
			cpu /= 100
		}
		out = append(out, containerMetricRow{
			ServiceID:    sid,
			Name:         names[sid],
			ProjectLabel: d.serviceProjectLabel(r.Context(), sid),
			TS:           nullInt64(row.Ts),
			CPUPercent:   cpu,
			MemoryBytes:  nullInt64(row.Mem),
			MemoryLimit:  nullInt64(row.MemLimit),
			NetRx:        nullInt64(row.NetRx),
			NetTx:        nullInt64(row.NetTx),
		})
	}
	_ = since
	writeJSON(w, http.StatusOK, map[string]any{
		"range":   rangeQ,
		"metrics": out,
	})
}

func (d apiDeps) enrichContainerRows(ctx context.Context, rows []docker.ContainerRow) []containerAPIRow {
	metrics, _ := d.q.ListLatestContainerMetrics(ctx, sqlc.ListLatestContainerMetricsParams{
		Res:   sql.NullInt64{Int64: eventlog.ResMinute, Valid: true},
		Res_2: sql.NullInt64{Int64: eventlog.ResMinute, Valid: true},
	})
	metricByService := make(map[string]sqlc.ListLatestContainerMetricsRow, len(metrics))
	for _, m := range metrics {
		if m.ServiceID.Valid {
			metricByService[m.ServiceID.String] = m
		}
	}
	out := make([]containerAPIRow, 0, len(rows))
	for _, row := range rows {
		api := containerAPIRow{
			ID:        row.ID,
			Name:      row.Name,
			State:     row.State,
			Kind:      row.Kind,
			ServiceID: row.ServiceID,
			CreatedAt: row.CreatedAt,
		}
		if row.ServiceID != "" {
			api.ProjectLabel = d.serviceProjectLabel(ctx, row.ServiceID)
			if m, ok := metricByService[row.ServiceID]; ok {
				cpu := nullFloat64(m.Cpu)
				if cpu > 1 {
					cpu /= 100
				}
				api.CPUPercent = cpu
				api.MemoryBytes = nullInt64(m.Mem)
				api.MemoryLimit = nullInt64(m.MemLimit)
			}
		} else if row.Kind == "system" {
			api.ProjectLabel = "Dockyard"
		}
		out = append(out, api)
	}
	return out
}

func (d apiDeps) serviceProjectLabel(ctx context.Context, serviceID string) string {
	if serviceID == "" || d.q == nil {
		return ""
	}
	row, err := d.q.GetServiceAnalyticsLabel(ctx, serviceID)
	if err != nil {
		return ""
	}
	if row.ProjectName != "" && row.EnvironmentName != "" {
		return fmt.Sprintf("%s / %s", row.ProjectName, row.EnvironmentName)
	}
	if row.ProjectSlug != "" {
		return row.ProjectSlug
	}
	return row.ProjectName
}

func (d apiDeps) containerNameByService(ctx context.Context) map[string]string {
	out := make(map[string]string)
	if d.docker == nil {
		return out
	}
	rows, err := d.docker.ListContainerRows(ctx)
	if err != nil {
		return out
	}
	for _, r := range rows {
		if r.ServiceID != "" {
			out[r.ServiceID] = r.Name
		}
	}
	return out
}

func metricRangeSince(rangeQ string) int64 {
	switch rangeQ {
	case "30d":
		return time.Now().UTC().Add(-30 * 24 * time.Hour).Unix()
	case "7d":
		return time.Now().UTC().Add(-7 * 24 * time.Hour).Unix()
	case "1h":
		return time.Now().UTC().Add(-time.Hour).Unix()
	default:
		return time.Now().UTC().Add(-24 * time.Hour).Unix()
	}
}
