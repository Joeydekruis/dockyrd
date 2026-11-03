package build

import (
	"errors"
	"fmt"
)

const defaultMinFreePercent = 10

// ErrDiskLow is returned when free disk on the data volume is below the guard threshold.
var ErrDiskLow = errors.New("insufficient disk space for build (free at least 10% on data volume)")

// CheckDiskGuard refuses builds when free space on the data directory filesystem is too low.
func CheckDiskGuard(dataDir string, minFreePercent int) error {
	if minFreePercent <= 0 {
		minFreePercent = defaultMinFreePercent
	}
	free, total, err := diskFreeRatio(dataDir)
	if err != nil {
		return nil
	}
	if total == 0 {
		return nil
	}
	pct := (free * 100) / total
	if pct < uint64(minFreePercent) {
		return fmt.Errorf("%w: %d%% free", ErrDiskLow, pct)
	}
	return nil
}
