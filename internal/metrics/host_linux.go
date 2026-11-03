//go:build linux

package metrics

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// SampleHost reads CPU, memory, disk, and load from /proc (Linux).
func SampleHost(dataDir string) (HostSample, error) {
	var s HostSample
	memUsed, memTotal, err := readMemInfo()
	if err == nil {
		s.MemUsed = memUsed
		s.MemTotal = memTotal
	}
	if cpu, err := readCPUUsage(); err == nil {
		s.CPU = cpu
	}
	if load, err := readLoad(); err == nil {
		s.Load1 = load
	}
	if dataDir != "" {
		if du, dt, err := diskUsage(dataDir); err == nil {
			s.DiskUsed = du
			s.DiskTotal = dt
		}
	}
	return s, nil
}

func readMemInfo() (used, total int64, err error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	var avail, free, buffers, cached int64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseInt(fields[1], 10, 64)
		val *= 1024
		switch fields[0] {
		case "MemTotal:":
			total = val
		case "MemAvailable:":
			avail = val
		case "MemFree:":
			free = val
		case "Buffers:":
			buffers = val
		case "Cached:":
			cached = val
		}
	}
	if avail > 0 {
		used = total - avail
	} else if total > 0 {
		used = total - free - buffers - cached
	}
	return used, total, sc.Err()
}

func readCPUUsage() (float64, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return 0, sc.Err()
	}
	fields := strings.Fields(sc.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, nil
	}
	var total, idle int64
	for i := 1; i < len(fields); i++ {
		v, _ := strconv.ParseInt(fields[i], 10, 64)
		total += v
		if i == 4 {
			idle = v
		}
	}
	if total == 0 {
		return 0, nil
	}
	return float64(total-idle) / float64(total) * 100, nil
}

func readLoad() (float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, nil
	}
	return strconv.ParseFloat(fields[0], 64)
}

func diskUsage(path string) (used, total int64, err error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0, err
	}
	total = int64(st.Blocks) * int64(st.Bsize)
	free := int64(st.Bfree) * int64(st.Bsize)
	return total - free, total, nil
}
