package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
	writeJSON(w, http.StatusOK, buildServiceConnectionResponse(d.cfg.PublicIPv4, row.Kind, row.Slug, raw, key, reveal))
}

func buildServiceConnectionResponse(publicHost, kind, slug string, raw json.RawMessage, key []byte, reveal bool) map[string]any {
	switch kind {
	case "postgres":
		pg, pw, err := dbkind.ParsePostgres(raw, key)
		if err != nil {
			return map[string]any{"variables": map[string]string{}}
		}
		internal := dbkind.Published("postgres", slug, pw, pg)
		maskConnectionSecrets(internal, reveal)
		external := map[string]any{
			"enabled":   pg.PublicPort > 0,
			"host":      "",
			"port":      pg.PublicPort,
			"variables": map[string]string{},
		}
		if pg.PublicPort > 0 {
			host := externalHost(publicHost)
			extVars := dbkind.PublishedPostgresAt(host, strconv.Itoa(pg.PublicPort), pw, pg)
			maskConnectionSecrets(extVars, reveal)
			external = map[string]any{
				"enabled":   true,
				"host":      host,
				"port":      pg.PublicPort,
				"variables": extVars,
			}
		}
		return map[string]any{
			"variables": internal,
			"internal":  map[string]any{"variables": internal},
			"external":  external,
		}
	case "redis":
		_, pw, err := dbkind.ParseRedis(raw, key)
		if err != nil {
			return map[string]any{"variables": map[string]string{}}
		}
		internal := dbkind.Published("redis", slug, pw, dbkind.PostgresConfig{})
		maskConnectionSecrets(internal, reveal)
		return map[string]any{
			"variables": internal,
			"internal":  map[string]any{"variables": internal},
			"external": map[string]any{
				"enabled":   false,
				"host":      "",
				"port":      0,
				"variables": map[string]string{},
			},
		}
	default:
		return map[string]any{"variables": map[string]string{}}
	}
}

func maskConnectionSecrets(vars map[string]string, reveal bool) {
	if reveal {
		return
	}
	for k := range vars {
		if strings.Contains(strings.ToUpper(k), "PASSWORD") || k == "DATABASE_URL" || k == "REDIS_URL" {
			vars[k] = "••••••••"
		}
	}
}
