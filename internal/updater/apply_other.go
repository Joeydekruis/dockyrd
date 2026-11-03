//go:build !linux

package updater

import "fmt"

func installBinary(_, _ string) error {
	return fmt.Errorf("self-update install is only supported on Linux hosts")
}

// RollbackBinary is a no-op off Linux.
func RollbackBinary(_ string) error {
	return fmt.Errorf("rollback is only supported on Linux hosts")
}
