package backup

import (
	"fmt"

	"github.com/Joeydekruis/dockyrd/internal/build"
)

// ErrDiskLow is returned when the data volume is too full to start a backup dump.
var ErrDiskLow = build.ErrDiskLow

// CheckDiskGuard refuses backup dumps when free space on the data directory is too low.
func CheckDiskGuard(dataDir string, minFreePercent int) error {
	if err := build.CheckDiskGuard(dataDir, minFreePercent); err != nil {
		return fmt.Errorf("backup skipped: %w", err)
	}
	return nil
}
