package resources

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func cgroupV2Unified() bool {
	_, err := os.Stat("/sys/fs/cgroup/cgroup.controllers")
	return err == nil
}

func sliceCgroupPath(slice string) string {
	// dy-build.slice → /sys/fs/cgroup/dy.slice/dy-build.slice
	parts := strings.Split(strings.TrimSuffix(slice, ".slice"), "-")
	if len(parts) < 2 {
		return filepath.Join("/sys/fs/cgroup", slice)
	}
	root := parts[0] + ".slice"
	return filepath.Join("/sys/fs/cgroup", root, slice)
}

func readCgroupInt64(path string) (int64, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(b))
	if s == "max" {
		return 0, fmt.Errorf("unlimited")
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func sliceWorkingSet(slice string) (int64, error) {
	base := sliceCgroupPath(slice)
	current, err := readCgroupInt64(filepath.Join(base, "memory.current"))
	if err != nil {
		return 0, err
	}
	inactiveFile, err := readMemoryStatKey(filepath.Join(base, "memory.stat"), "inactive_file")
	if err != nil {
		inactiveFile = 0
	}
	ws := current - inactiveFile
	if ws < 0 {
		ws = current
	}
	return ws, nil
}

func readMemoryStatKey(path, key string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 2 || fields[0] != key {
			continue
		}
		return strconv.ParseInt(fields[1], 10, 64)
	}
	return 0, os.ErrNotExist
}

func writeCgroupFreeze(slice string, frozen bool) error {
	val := "0"
	if frozen {
		val = "1"
	}
	return os.WriteFile(filepath.Join(sliceCgroupPath(slice), "cgroup.freeze"), []byte(val), 0o644)
}

func readBuildOOMKills() (uint64, error) {
	path := filepath.Join(sliceCgroupPath(SliceBuild), "memory.events")
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 2 && fields[0] == "oom_kill" {
			return strconv.ParseUint(fields[1], 10, 64)
		}
	}
	return 0, nil
}
