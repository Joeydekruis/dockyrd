package updater

import (
	"fmt"
	"strings"
)

// releaseAssetPrefixes returns canonical dockyrd_* then legacy dockyard_* names.
func releaseAssetPrefixes(version, goos, goarch string) []string {
	ver := strings.TrimPrefix(strings.TrimSpace(version), "v")
	return []string{
		fmt.Sprintf("dockyrd_%s_%s_%s", ver, goos, goarch),
		fmt.Sprintf("dockyard_%s_%s_%s", ver, goos, goarch),
	}
}
