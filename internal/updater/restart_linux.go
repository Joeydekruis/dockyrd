//go:build linux

package updater

import (
	"os"
	"os/exec"
	"time"
)

// ScheduleServiceRestart restarts the systemd unit after a short delay so HTTP responses can flush.
func ScheduleServiceRestart() {
	unit := os.Getenv("dockyrd_SYSTEMD_UNIT")
	if unit == "" {
		unit = os.Getenv("DOCKYARD_SYSTEMD_UNIT")
	}
	if unit == "" {
		unit = "dockyrd.service"
	}
	go func() {
		time.Sleep(750 * time.Millisecond)
		//nolint:gosec // trusted fixed argv; unit name from env with default
		_ = exec.Command("systemctl", "restart", unit).Run()
	}()
}
