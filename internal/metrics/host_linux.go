//go:build linux

package metrics

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var cpuPrev struct {
	sync.Mutex
	total int64
	idle  int64
	ok    bool
}

var corePrev struct {
	sync.Mutex
	total map[int]int64
	idle  map[int]int64
	ok    bool
}

// SampleHost reads CPU, memory, disk, and load from /proc (Linux).
func SampleHost(dataDir string) (HostSample, error) {
	var s HostSample
	memUsed, memTotal, err := readMemInfo()
	if err == nil {
		s.MemUsed = memUsed
		s.MemTotal = memTotal
	}
	if cpu, ok := readCPUUsageFraction(); ok {
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
	if rx, tx, err := readNetBytes(); err == nil {
		s.NetRx = rx
		s.NetTx = tx
	}
	if swapUsed, swapTotal, err := readSwap(); err == nil {
		s.SwapUsed = swapUsed
		s.SwapTotal = swapTotal
	}
	s.CoreUsage = readPerCoreCPUFractions()
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

func readSwap() (used, total int64, err error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	var swapFree int64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseInt(fields[1], 10, 64)
		val *= 1024
		switch fields[0] {
		case "SwapTotal:":
			total = val
		case "SwapFree:":
			swapFree = val
		}
	}
	if total > 0 {
		used = total - swapFree
	}
	return used, total, sc.Err()
}

func readPerCoreCPUFractions() []float64 {
	counters, err := readPerCoreCounters()
	if err != nil || len(counters) == 0 {
		return nil
	}
	corePrev.Lock()
	defer corePrev.Unlock()
	out := make([]float64, 0, len(counters))
	if !corePrev.ok {
		corePrev.total = make(map[int]int64, len(counters))
		corePrev.idle = make(map[int]int64, len(counters))
		for idx, c := range counters {
			corePrev.total[idx] = c.total
			corePrev.idle[idx] = c.idle
			out = append(out, 0)
		}
		corePrev.ok = true
		return out
	}
	for idx, c := range counters {
		pt := corePrev.total[idx]
		pi := corePrev.idle[idx]
		corePrev.total[idx] = c.total
		corePrev.idle[idx] = c.idle
		dTotal := c.total - pt
		dIdle := c.idle - pi
		busy := 0.0
		if dTotal > 0 {
			busy = 1 - float64(dIdle)/float64(dTotal)
		}
		if busy < 0 {
			busy = 0
		}
		if busy > 1 {
			busy = 1
		}
		out = append(out, busy)
	}
	return out
}

type coreCounter struct {
	total int64
	idle  int64
}

func readPerCoreCounters() (map[int]coreCounter, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := make(map[int]coreCounter)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 5 || !strings.HasPrefix(fields[0], "cpu") || fields[0] == "cpu" {
			continue
		}
		idStr := strings.TrimPrefix(fields[0], "cpu")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		var total, idle int64
		for i := 1; i < len(fields); i++ {
			v, _ := strconv.ParseInt(fields[i], 10, 64)
			total += v
			if i == 4 {
				idle = v
			}
		}
		out[id] = coreCounter{total: total, idle: idle}
	}
	return out, sc.Err()
}

// readCPUUsageFraction returns busy CPU in [0,1] from delta since the prior sample.
func readCPUUsageFraction() (float64, bool) {
	total, idle, err := readCPUCounters()
	if err != nil || total <= 0 {
		return 0, false
	}
	cpuPrev.Lock()
	defer cpuPrev.Unlock()
	if !cpuPrev.ok {
		cpuPrev.total = total
		cpuPrev.idle = idle
		cpuPrev.ok = true
		return 0, true
	}
	dTotal := total - cpuPrev.total
	dIdle := idle - cpuPrev.idle
	cpuPrev.total = total
	cpuPrev.idle = idle
	if dTotal <= 0 {
		return 0, true
	}
	busy := 1 - float64(dIdle)/float64(dTotal)
	if busy < 0 {
		busy = 0
	}
	if busy > 1 {
		busy = 1
	}
	return busy, true
}

func readCPUCounters() (total, idle int64, err error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return 0, 0, sc.Err()
	}
	fields := strings.Fields(sc.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, 0, nil
	}
	for i := 1; i < len(fields); i++ {
		v, _ := strconv.ParseInt(fields[i], 10, 64)
		total += v
		if i == 4 {
			idle = v
		}
	}
	return total, idle, sc.Err()
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

func readNetBytes() (rx, tx int64, err error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "Inter-|") || strings.HasPrefix(line, " face |") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}
		r, _ := strconv.ParseInt(fields[0], 10, 64)
		t, _ := strconv.ParseInt(fields[8], 10, 64)
		rx += r
		tx += t
	}
	return rx, tx, sc.Err()
}
