//go:build unix

package diskusage

import "syscall"

func sameFilesystem(a, b string) (bool, error) {
	if a == b {
		return true, nil
	}
	var sa, sb syscall.Statfs_t
	if err := syscall.Statfs(a, &sa); err != nil {
		return false, err
	}
	if err := syscall.Statfs(b, &sb); err != nil {
		return false, err
	}
	return sa.Fsid == sb.Fsid, nil
}
