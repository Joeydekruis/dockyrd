package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/updater"
)

type serverInfoResponse struct {
	Version         string `json:"version"`
	UptimeSec       int64  `json:"uptime_sec"`
	UpdateAvailable bool   `json:"update_available"`
	LatestTag       string `json:"latest_tag,omitempty"`
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
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		check, err := updater.Check(ctx, BuildVersion(), updater.Options{})
		if err == nil {
			resp.UpdateAvailable = check.Newer
			if check.Newer {
				resp.LatestTag = check.Latest.Tag
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
