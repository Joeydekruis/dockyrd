package resources

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// PSI holds memory pressure averages (avg10) from cgroup PSI files.
type PSI struct {
	AppsSomeAvg10  float64
	SystemFullAvg10 float64
}

// ReadPSI reads app slice and system memory pressure (best effort).
func ReadPSI() PSI {
	var out PSI
	appsPath := filepath.Join(sliceCgroupPath(SliceApps), "memory.pressure")
	out.AppsSomeAvg10, _ = readPSIAvg10(appsPath, "some")
	out.SystemFullAvg10, _ = readPSIAvg10("/proc/pressure/memory", "full")
	return out
}

func readPSIAvg10(path, kind string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	prefix := kind + " avg10="
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.Contains(line, prefix) {
			continue
		}
		idx := strings.Index(line, prefix)
		rest := line[idx+len(prefix):]
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		return strconv.ParseFloat(fields[0], 64)
	}
	return 0, os.ErrNotExist
}
