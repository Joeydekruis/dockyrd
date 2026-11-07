//go:build !linux

package updater

func linkPanelBinaryAliases(_ string) error {
	return nil
}
