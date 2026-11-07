package server

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Joeydekruis/dockyrd/internal/hostops"
)

func registerSecurityHostActionRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("POST /api/v1/security/host-actions/{id}/confirm", deps.handleHostActionConfirm)
	mux.HandleFunc("POST /api/v1/security/host-actions/{id}/revert", deps.handleHostActionRevert)
}

func (d apiDeps) hostRunner() (*hostops.Runner, error) {
	if d.q == nil || d.db == nil {
		return nil, errors.New("database not ready")
	}
	reg := hostops.DefaultRegistry(d.cfg)
	timer := hostops.NewSystemdLockoutTimer()
	return hostops.NewRunner(d.db, d.q, reg, d.audit, timer), nil
}

func (d apiDeps) handleHostActionConfirm(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id required")
		return
	}
	runner, err := d.hostRunner()
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if err := runner.Confirm(r.Context(), id, user.ID); err != nil {
		if errors.Is(err, hostops.ErrInvalidTransition) {
			writeJSONError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "host action not found")
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id, "state": hostops.StateConfirmed})
}

func (d apiDeps) handleHostActionRevert(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "id required")
		return
	}
	runner, err := d.hostRunner()
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	meta := hostops.RevertMeta{Source: "api", Actor: user.ID}
	if err := runner.Revert(r.Context(), id, meta); err != nil {
		if errors.Is(err, hostops.ErrInvalidTransition) {
			writeJSONError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "host action not found")
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": id, "state": hostops.StateReverted})
}
