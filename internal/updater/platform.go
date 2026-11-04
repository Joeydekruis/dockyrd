package updater

import "runtime"

// InstallSupported reports whether self-update can replace the host binary on this OS.
func InstallSupported() bool {
	return runtime.GOOS == "linux"
}
