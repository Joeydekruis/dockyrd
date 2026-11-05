package docker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const buildkitDataVolume = "dy-buildkit-data"

// BuildkitConfigDir returns the host directory for buildkitd.toml.
func BuildkitConfigDir(dataDir string) string {
	return filepath.Join(dataDir, "buildkit")
}

// RenderBuildkitToml writes buildkitd.toml sized to this host.
func RenderBuildkitToml(dataDir string) (string, error) {
	dir := BuildkitConfigDir(dataDir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "buildkitd.toml")
	parallelism := buildkitParallelism()
	maxUsed, minFree := buildkitDiskLimits(dataDir)
	body := fmt.Sprintf(`[worker.oci]
  max-parallelism = %d
  gc = true
  reservedSpace = "2GB"
  maxUsedSpace = "%s"
  minFreeSpace = "%s"

[worker.containerd]
  enabled = false
`, parallelism, maxUsed, minFree)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func buildkitParallelism() int {
	memGB := hostMemTotalGB()
	switch {
	case memGB < 4:
		return 1
	case memGB <= 8:
		return 2
	default:
		n := runtime.NumCPU()
		if n > 4 {
			n = 4
		}
		if n < 1 {
			n = 1
		}
		return n
	}
}

func hostMemTotalGB() int64 {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 4
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "MemTotal:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			break
		}
		kb, _ := strconv.ParseInt(fields[1], 10, 64)
		return kb / (1024 * 1024)
	}
	return 4
}

func buildkitDiskLimits(dataDir string) (maxUsed, minFree string) {
	free, total, err := statfsBytes(dataDir)
	if err != nil || total == 0 {
		return "30GB", "10%"
	}
	maxBytes := (total * 20) / 100
	const cap = 30 * 1024 * 1024 * 1024
	if maxBytes > cap {
		maxBytes = cap
	}
	minPct := (free * 100) / total
	if minPct < 10 {
		minPct = 10
	}
	gb := maxBytes / (1024 * 1024 * 1024)
	if gb < 1 {
		gb = 1
	}
	return fmt.Sprintf("%dGB", gb), fmt.Sprintf("%d%%", minPct)
}

func statfsBytes(path string) (free, total uint64, err error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0, err
	}
	free = st.Bavail * uint64(st.Bsize)
	total = st.Blocks * uint64(st.Bsize)
	return free, total, nil
}
