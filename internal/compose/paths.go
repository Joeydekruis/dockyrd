package compose

import "path/filepath"

// StackDataDir is /var/lib/dockyrd/compose/<service_id>/.
func StackDataDir(dataDir, serviceID string) string {
	return filepath.Join(dataDir, "compose", serviceID)
}
