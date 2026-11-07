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
	CheckedAt        int64  `json:"checked_at,omitempty"`
	FromCache        bool   `json:"from_cache"`
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
	refresh := r.URL.Query().Get("refresh") == "1"
	writeJSON(w, http.StatusOK, d.buildUpdatesStatus(r.Context(), refresh))
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
		Repo:   os.Getenv("dockyrd_GITHUB_REPO"),
		Quiet:  true,
		DryRun: os.Getenv("dockyrd_UPDATE_DRY_RUN") == "1",
	}
	if err := updater.DefaultReleaseCache.RefreshForce(ctx, BuildVersion(), opts); err != nil {
		// Apply still re-checks; continue with best-effort cache refresh.
		_ = err
	}
	result, err := updater.Apply(ctx, d.cfg, BuildVersion(), opts)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "dockyrd.update.apply",
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
		msg = "Update installed. Restarting dockyrd…"
		updater.ScheduleServiceRestart()
	}

	writeJSON(w, http.StatusOK, applyUpdateResponse{
		OK:           true,
		InstalledTag: result.InstalledTag,
		Message:      msg,
		Restarting:   restarting,
	})
}

func (d apiDeps) buildUpdatesStatus(ctx context.Context, manualRefresh bool) updatesStatusResponse {
	resp := updatesStatusResponse{
		Version:        BuildVersion(),
		ApplySupported: updater.InstallSupported(),
	}
	if !resp.ApplySupported {
		resp.ApplyUnsupported = "Self-update install runs on the Linux host binary only."
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := updater.Options{Repo: os.Getenv("dockyrd_GITHUB_REPO")}
	var snap updater.ReleaseSnapshot
	if manualRefresh {
		snap = updater.DefaultReleaseCache.RefreshManual(ctx, BuildVersion(), opts)
	} else {
		snap = updater.DefaultReleaseCache.Snapshot()
		if !snap.HasData {
			resp.CheckError = "Waiting for the first scheduled update check (every 5 minutes)."
			return resp
		}
	}

	resp.FromCache = snap.FromCache
	if !snap.CheckedAt.IsZero() {
		resp.CheckedAt = snap.CheckedAt.Unix()
	}
	if snap.Err != nil {
		resp.CheckError = snap.Err.Error()
		if snap.HasData && snap.Result.Latest.Tag != "" {
			resp.CheckOK = true
			resp.UpdateAvailable = snap.Result.Newer
			if snap.Result.Newer {
				resp.LatestTag = snap.Result.Latest.Tag
				resp.ReleaseNotes = snap.Result.Latest.Notes
			}
		}
		return resp
	}
	resp.CheckOK = true
	resp.UpdateAvailable = snap.Result.Newer
	if snap.Result.Newer {
		resp.LatestTag = snap.Result.Latest.Tag
		resp.ReleaseNotes = snap.Result.Latest.Notes
	}
	return resp
}
