package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/reconcile"
	"github.com/Joeydekruis/dockyrd/internal/serviceops"
)

func registerDatabaseServiceRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("PATCH /api/v1/services/{id}/public-access", deps.handlePatchDatabasePublicAccess)
}

func (d apiDeps) handlePatchDatabasePublicAccess(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	var body struct {
		Enabled bool `json:"enabled"`
		Port    int  `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
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
	if row.Kind != "postgres" {
		writeJSONError(w, http.StatusBadRequest, "public access applies to postgres only")
		return
	}
	raw := serviceConfigJSON(row.Config)
	port := 0
	if body.Enabled {
		port = body.Port
		if port <= 0 {
			port = pickPublicPort(row.ID)
		}
		if port < 1024 || port > 65535 {
			writeJSONError(w, http.StatusBadRequest, "port must be between 1024 and 65535")
			return
		}
	}
	merged, err := dbkind.SetPostgresPublicPort(raw, port)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	now := time.Now().Unix()
	if err := d.q.UpdateServiceConfig(ctx, sqlc.UpdateServiceConfigParams{
		Config:    merged,
		UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
		ID:        id,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update error")
		return
	}
	_ = serviceops.StopManagedDatabaseContainer(ctx, d.docker, id)
	reconcile.SyncDatabaseServices(ctx, d.cfg, d.q, d.docker, nil)
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "database.public_access",
		TargetKind: "service",
		TargetID:   id,
		Meta:       map[string]any{"enabled": body.Enabled, "port": port},
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"enabled": port > 0,
		"port":    port,
	})
}

func pickPublicPort(serviceID string) int {
	var sum int
	for i := 0; i < len(serviceID); i++ {
		sum += int(serviceID[i])
	}
	return 40000 + (sum % 10000)
}

func externalHost(cfgHost string) string {
	if h := cfgHost; h != "" {
		return h
	}
	return "your-server-ip"
}
