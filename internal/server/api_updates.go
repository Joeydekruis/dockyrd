package server

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/updater"
)

var updateApplyMu sync.Mutex

type updatesStatusResponse struct {
	Version          string `json:"version"`
	UpdateAvailable  bool   `json:"update_available"`
	LatestTag        string `json:"latest_tag,omitempty"`
	ReleaseNotes     string `json:"release_notes,omitempty"`
	CheckOK          bool   `json:"check_ok"`
	CheckError       string `json:"check_error,omitempty"`
	ApplySupported   bool   `json:"apply_supported"`
	ApplyUnsupported string `json:"apply_unsupported_reason,omitempty"`
}

type applyUpdateResponse struct {
	OK           bool   `json:"ok"`
	InstalledTag string `json:"installed_tag,omitempty"`
	Message      string `json:"message"`
	Restarting   bool   `json:"restarting"`
}

func registerUpdatesRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/updates", deps.handleUpdatesGet)
	mux.HandleFunc("POST /api/v1/updates/apply", deps.handleUpdatesApply)
}

func (d apiDeps) handleUpdatesGet(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	writeJSON(w, http.StatusOK, d.buildUpdatesStatus(r.Context()))
}

func (d apiDeps) handleUpdatesApply(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	if !updateApplyMu.TryLock() {
		writeJSONError(w, http.StatusConflict, "update already in progress")
		return
	}
	defer updateApplyMu.Unlock()

	if !updater.InstallSupported() {
		writeJSONError(w, http.StatusNotImplemented, "self-update apply is only supported on Linux production hosts")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 12*time.Minute)
	defer cancel()

	opts := updater.Options{
		Repo:   os.Getenv("DOCKYARD_GITHUB_REPO"),
		Quiet:  true,
		DryRun: os.Getenv("DOCKYARD_UPDATE_DRY_RUN") == "1",
	}
	result, err := updater.Apply(ctx, d.cfg, BuildVersion(), opts)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "dockyard.update.apply",
		TargetKind: "server",
		Meta:       map[string]any{"from": BuildVersion(), "to": result.InstalledTag},
		IP:         clientIP(r),
	})

	restarting := false
	msg := "Update installed."
	if opts.DryRun {
		msg = "Dry run: update would be installed."
	} else {
		restarting = true
		msg = "Update installed. Restarting Dockyard…"
		updater.ScheduleServiceRestart()
	}

	writeJSON(w, http.StatusOK, applyUpdateResponse{
		OK:           true,
		InstalledTag: result.InstalledTag,
		Message:      msg,
		Restarting:   restarting,
	})
}

func (d apiDeps) buildUpdatesStatus(ctx context.Context) updatesStatusResponse {
	resp := updatesStatusResponse{
		Version:        BuildVersion(),
		ApplySupported: updater.InstallSupported(),
	}
	if !resp.ApplySupported {
		resp.ApplyUnsupported = "Self-update install runs on the Linux host binary only."
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	check, err := updater.Check(ctx, BuildVersion(), updater.Options{Repo: os.Getenv("DOCKYARD_GITHUB_REPO")})
	if err != nil {
		resp.CheckError = err.Error()
		return resp
	}
	resp.CheckOK = true
	resp.UpdateAvailable = check.Newer
	if check.Newer {
		resp.LatestTag = check.Latest.Tag
		resp.ReleaseNotes = check.Latest.Notes
	}
	return resp
}
