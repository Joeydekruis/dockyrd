package diskusage

import (
	"os"
	"path/filepath"
)

// DirSize sums regular file sizes under root (directories are not counted).
func DirSize(root string) int64 {
	info, err := os.Stat(root)
	if err != nil {
		return 0
	}
	if !info.IsDir() {
		return info.Size()
	}
	var n int64
	_ = filepath.WalkDir(root, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return nil
		}
		n += fi.Size()
		return nil
	})
	return n
}

// FileSize returns the size of a single file, or 0 when missing.
func FileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return 0
	}
	return info.Size()
}
