package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
)

func registerConnectionRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/connection", deps.handleServiceConnection)
}

func (d apiDeps) handleServiceConnection(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	row, err := d.q.GetServiceByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if row.Kind != "postgres" && row.Kind != "redis" {
		writeJSONError(w, http.StatusBadRequest, "connection info only for postgres or redis")
		return
	}
	reveal := r.URL.Query().Get("reveal") == "1"
	key, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "secrets not configured")
		return
	}
	raw := serviceConfigJSON(row.Config)
	vars := map[string]string{}
	switch row.Kind {
	case "postgres":
		pg, pw, err := dbkind.ParsePostgres(raw, key)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "config error")
			return
		}
		vars = dbkind.Published("postgres", row.Slug, pw, pg)
	case "redis":
		_, pw, err := dbkind.ParseRedis(raw, key)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "config error")
			return
		}
		vars = dbkind.Published("redis", row.Slug, pw, dbkind.PostgresConfig{})
	}
	if !reveal {
		for k := range vars {
			if strings.Contains(strings.ToUpper(k), "PASSWORD") || k == "DATABASE_URL" || k == "REDIS_URL" {
				vars[k] = "••••••••"
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"variables": vars})
}
