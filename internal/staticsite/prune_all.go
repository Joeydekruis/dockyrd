package staticsite

import (
	"os"
	"path/filepath"
	"sort"
)

const DefaultReleaseKeep = 5

// ReclaimableReleaseBytes sums on-disk size of release dirs that PruneReleases would remove.
func ReclaimableReleaseBytes(sitesRoot string, keep int) int64 {
	if keep <= 0 {
		keep = DefaultReleaseKeep
	}
	entries, err := os.ReadDir(sitesRoot)
	if err != nil {
		return 0
	}
	var total int64
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		total += reclaimableForService(sitesRoot, e.Name(), keep)
	}
	return total
}

// PruneAllOldReleases runs release retention for every service under sitesRoot.
func PruneAllOldReleases(sitesRoot string, keep int) int64 {
	if keep <= 0 {
		keep = DefaultReleaseKeep
	}
	before := ReclaimableReleaseBytes(sitesRoot, keep)
	entries, err := os.ReadDir(sitesRoot)
	if err != nil {
		return 0
	}
	for _, e := range entries {
		if e.IsDir() {
			PruneReleases(sitesRoot, e.Name(), keep)
		}
	}
	after := ReclaimableReleaseBytes(sitesRoot, keep)
	if before > after {
		return before - after
	}
	return before
}

func reclaimableForService(sitesRoot, serviceSlug string, keep int) int64 {
	paths := PathsForService(sitesRoot, serviceSlug)
	entries, err := os.ReadDir(paths.ReleasesDir)
	if err != nil {
		return 0
	}
	type item struct {
		name string
		mod  int64
	}
	var dirs []item
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		dirs = append(dirs, item{name: e.Name(), mod: info.ModTime().Unix()})
	}
	if len(dirs) <= keep {
		return 0
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].mod > dirs[j].mod })
	var total int64
	for _, d := range dirs[keep:] {
		total += dirSize(filepath.Join(paths.ReleasesDir, d.name))
	}
	return total
}

func dirSize(root string) int64 {
	var n int64
	_ = filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		n += info.Size()
		return nil
	})
	return n
}
