//go:build linux

package updater

import (
	"os"
	"path/filepath"
)

const (
	legacyPanelBin = "/usr/local/bin/dockyard"
	canonPanelBin  = "/usr/local/bin/dockyrd"
)

// linkPanelBinaryAliases keeps dockyard ↔ dockyrd symlinks in sync after self-update
// so older systemd units that ExecStart=/usr/local/bin/dockyard keep working.
func linkPanelBinaryAliases(installedPath string) error {
	installedPath = filepath.Clean(installedPath)
	switch installedPath {
	case canonPanelBin:
		return symlinkForce(canonPanelBin, legacyPanelBin)
	case legacyPanelBin:
		return symlinkForce(legacyPanelBin, canonPanelBin)
	default:
		return nil
	}
}

func symlinkForce(target, linkPath string) error {
	_ = os.Remove(linkPath)
	return os.Symlink(target, linkPath)
}
