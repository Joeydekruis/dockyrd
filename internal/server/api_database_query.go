package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/dbquery"
)

func registerDatabaseQueryRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("POST /api/v1/services/{id}/query", deps.handleDatabaseQuery)
	mux.HandleFunc("GET /api/v1/services/{id}/schema", deps.handleDatabaseSchema)
}

func (d apiDeps) handleDatabaseQuery(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	serviceID := r.PathValue("id")
	var body struct {
		SQL     string `json:"sql"`
		MaxRows int    `json:"max_rows"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
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
	if svc.Kind != "postgres" {
		writeJSONError(w, http.StatusBadRequest, "query studio is available for postgres only")
		return
	}
	result, err := dbquery.RunPostgres(ctx, d.cfg, d.docker, svc, body.SQL, body.MaxRows)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	prefix := strings.TrimSpace(body.SQL)
	if len(prefix) > 80 {
		prefix = prefix[:80]
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "database.query",
		TargetKind: "service",
		TargetID:   serviceID,
		Meta: map[string]any{
			"row_count":  result.RowCount,
			"sql_prefix": prefix,
		},
	})
	writeJSON(w, http.StatusOK, result)
}

func (d apiDeps) handleDatabaseSchema(w http.ResponseWriter, r *http.Request) {
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
	if svc.Kind != "postgres" {
		writeJSONError(w, http.StatusBadRequest, "schema browser is available for postgres only")
		return
	}
	tables, err := dbquery.ListPostgresSchema(ctx, d.cfg, d.docker, svc)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tables": tables})
}
