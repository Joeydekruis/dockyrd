//go:build linux

package updater

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func installBinary(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	prev := dest + ".prev"
	if _, err := os.Stat(dest); err == nil {
		if err := copyFile(dest, prev); err != nil {
			return fmt.Errorf("backup previous binary: %w", err)
		}
	}
	tmp := dest + ".new"
	if err := copyFile(src, tmp); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		return err
	}
	return nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// RollbackBinary restores dest from dest.prev when present.
func RollbackBinary(dest string) error {
	prev := dest + ".prev"
	if _, err := os.Stat(prev); err != nil {
		return fmt.Errorf("no previous binary at %s", prev)
	}
	if err := copyFile(prev, dest); err != nil {
		return err
	}
	return os.Chmod(dest, 0o755)
}
