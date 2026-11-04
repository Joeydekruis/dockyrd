package server

import (
	"net/http"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/updater"
)

type serverInfoResponse struct {
	Version         string `json:"version"`
	UptimeSec       int64  `json:"uptime_sec"`
	UpdateAvailable bool   `json:"update_available"`
	LatestTag       string `json:"latest_tag,omitempty"`
	ReleaseNotes    string `json:"release_notes,omitempty"`
}

func (d apiDeps) handleServerInfo(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	resp := serverInfoResponse{
		Version:   BuildVersion(),
		UptimeSec: time.Now().Unix() - ProcessStartedAt(),
	}
	if user.Role == "owner" || user.Role == "admin" {
		snap := updater.DefaultReleaseCache.Snapshot()
		if snap.HasData && snap.Err == nil {
			resp.UpdateAvailable = snap.Result.Newer
			if snap.Result.Newer {
				resp.LatestTag = snap.Result.Latest.Tag
				resp.ReleaseNotes = snap.Result.Latest.Notes
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
