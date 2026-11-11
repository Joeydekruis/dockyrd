package server

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/diskusage"
)

const diskUsageCacheTTL = 2 * time.Minute

var (
	diskUsageMu    sync.Mutex
	diskUsageCache diskusage.Report
	diskUsageAt    time.Time
)

func (d apiDeps) handleServerDiskUsage(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	force := strings.TrimSpace(r.URL.Query().Get("refresh")) == "1"
	report := d.diskUsageReport(r, force)
	writeJSON(w, http.StatusOK, report)
}

func (d apiDeps) diskUsageReport(r *http.Request, force bool) diskusage.Report {
	if !force {
		diskUsageMu.Lock()
		if time.Since(diskUsageAt) < diskUsageCacheTTL && diskUsageCache.DataDir != "" {
			cached := diskUsageCache
			diskUsageMu.Unlock()
			return cached
		}
		diskUsageMu.Unlock()
	}
	report := diskusage.BuildReport(r.Context(), d.cfg, d.q, d.docker)
	report.ScannedAt = time.Now().UTC().Unix()
	diskUsageMu.Lock()
	diskUsageCache = report
	diskUsageAt = time.Now()
	diskUsageMu.Unlock()
	return report
}
