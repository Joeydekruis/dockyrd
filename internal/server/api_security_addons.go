package server

import (
	"errors"
	"net/http"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/hostops"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
	"github.com/Joeydekruis/dockyrd/internal/security/addons"
)

func registerSecurityAddonRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/security/addons", deps.handleSecurityAddonsList)
	mux.HandleFunc("POST /api/v1/security/addons/{id}/plan", deps.handleSecurityAddonPlan)
	mux.HandleFunc("POST /api/v1/security/addons/{id}/install", deps.handleSecurityAddonInstall)
	mux.HandleFunc("POST /api/v1/security/addons/{id}/uninstall", deps.handleSecurityAddonUninstall)
}

func (d apiDeps) handleSecurityAddonsList(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	rows, err := addons.ListMerged(r.Context(), d.q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"addons": rows})
}

func (d apiDeps) handleSecurityAddonPlan(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	reg := hostops.DefaultRegistry(d.cfg)
	plan, err := addons.PlanInstall(r.Context(), reg, id)
	if err != nil {
		if errors.Is(err, addons.ErrNotInstallable) {
			writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if errors.Is(err, hostops.ErrPrecondition) {
			writeJSONError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "plan": plan})
}

func (d apiDeps) handleSecurityAddonInstall(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	entry, err := addons.ByID(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if !entry.Installable {
		writeJSONError(w, http.StatusUnprocessableEntity, addons.ErrNotInstallable.Error())
		return
	}
	if err := addons.MarkInstalling(r.Context(), d.q, id); err != nil {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}
	jobID, err := jobs.EnqueueSecurityAddonInstall(r.Context(), d.q, id, user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = d.audit.Record(r.Context(), audit.Entry{
		UserID: user.ID, Action: "addon.installing", TargetKind: "addon", TargetID: id,
		Meta: map[string]any{"job_id": jobID},
	})
	writeJSON(w, http.StatusAccepted, map[string]any{"addon_id": id, "job_id": jobID, "state": "installing"})
}

func (d apiDeps) handleSecurityAddonUninstall(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if _, err := addons.ByID(id); err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := addons.MarkUninstalling(r.Context(), d.q, id); err != nil {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}
	jobID, err := jobs.EnqueueSecurityAddonUninstall(r.Context(), d.q, id, user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = d.audit.Record(r.Context(), audit.Entry{
		UserID: user.ID, Action: "addon.uninstalling", TargetKind: "addon", TargetID: id,
		Meta: map[string]any{"job_id": jobID},
	})
	writeJSON(w, http.StatusAccepted, map[string]any{"addon_id": id, "job_id": jobID, "state": "uninstalling"})
}
