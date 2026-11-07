package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/hostops"
	"github.com/Joeydekruis/dockyrd/internal/securitychecks"
)

func registerSecurityCheckRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/security/checks", deps.handleSecurityChecksGet)
	mux.HandleFunc("POST /api/v1/security/checks/run", deps.handleSecurityChecksRun)
	mux.HandleFunc("POST /api/v1/security/checks/{id}/fix", deps.handleSecurityCheckFix)
}

func (d apiDeps) securityDeps() securitychecks.Deps {
	return securitychecks.Deps{
		Cfg:    d.cfg,
		Q:      d.q,
		Docker: d.docker,
		Env:    hostops.NewHostEnv(d.cfg),
	}
}

func (d apiDeps) handleSecurityChecksGet(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	force := r.URL.Query().Get("refresh") == "1"
	if !force {
		if rep, ok, err := securitychecks.LoadCached(r.Context(), d.q, time.Now()); err == nil && ok {
			writeJSON(w, http.StatusOK, rep)
			return
		}
	}
	rep := securitychecks.RunAll(r.Context(), d.securityDeps())
	_ = securitychecks.Save(r.Context(), d.q, rep)
	writeJSON(w, http.StatusOK, rep)
}

func (d apiDeps) handleSecurityChecksRun(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	rep := securitychecks.RunAll(r.Context(), d.securityDeps())
	_ = securitychecks.Save(r.Context(), d.q, rep)
	_ = d.audit.Record(r.Context(), audit.Entry{
		UserID: user.ID, Action: "security.checks.run", TargetKind: "security", Meta: map[string]any{"ran_at": rep.RanAt},
	})
	writeJSON(w, http.StatusOK, rep)
}

func (d apiDeps) handleSecurityCheckFix(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	checkID := r.PathValue("id")
	kind, ok := securitychecks.FixKindForCheck(checkID)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "no host fix for check")
		return
	}
	runner, err := d.hostRunner()
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	actionID, err := runner.Plan(r.Context(), kind, user.ID)
	if err != nil {
		if errors.Is(err, hostops.ErrPrecondition) {
			writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := runner.Apply(r.Context(), actionID, user.ID); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := runner.Verify(r.Context(), actionID); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	row, _ := runner.Q.GetHostAction(r.Context(), actionID)
	_ = d.audit.Record(r.Context(), audit.Entry{
		UserID: user.ID, Action: "security.check.fix", TargetKind: "host_action", TargetID: actionID,
		Meta: map[string]any{"check_id": checkID, "kind": kind},
	})
	out := map[string]any{"host_action_id": actionID, "state": row.State}
	if row.ConfirmDeadline.Valid {
		out["confirm_deadline"] = row.ConfirmDeadline.Int64
	}
	writeJSON(w, http.StatusOK, out)
}
