package server

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/metrics"
	"github.com/Joeydekruis/dockyrd/internal/updater"
)

type coreUsageResponse struct {
	Core int     `json:"core"`
	CPU  float64 `json:"cpu"`
}

type serverInfoResponse struct {
	Version         string              `json:"version"`
	UptimeSec       int64               `json:"uptime_sec"`
	Hostname        string              `json:"hostname,omitempty"`
	PublicIPv4      string              `json:"public_ipv4,omitempty"`
	CPUCount        int             `json:"cpu_count,omitempty"`
	SwapUsed        int64               `json:"swap_used,omitempty"`
	SwapTotal       int64               `json:"swap_total,omitempty"`
	Cores           []coreUsageResponse `json:"cores,omitempty"`
	UpdateAvailable bool                `json:"update_available"`
	LatestTag       string              `json:"latest_tag,omitempty"`
	ReleaseNotes    string              `json:"release_notes,omitempty"`
}

func (d apiDeps) handleServerInfo(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	resp := serverInfoResponse{
		Version:    BuildVersion(),
		UptimeSec:  time.Now().Unix() - ProcessStartedAt(),
		CPUCount:   runtime.NumCPU(),
		PublicIPv4: d.cfg.EffectivePublicIPv4(),
	}
	if host, err := os.Hostname(); err == nil && host != "" {
		resp.Hostname = host
	}
	if sample, err := metrics.SampleHost(d.cfg.DataDir); err == nil {
		resp.SwapUsed = sample.SwapUsed
		resp.SwapTotal = sample.SwapTotal
		if len(sample.CoreUsage) > 0 {
			resp.Cores = make([]coreUsageResponse, len(sample.CoreUsage))
			for i, frac := range sample.CoreUsage {
				resp.Cores[i] = coreUsageResponse{Core: i + 1, CPU: frac}
			}
		}
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
