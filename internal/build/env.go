package build

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// InjectBuildEnv adds safe defaults when the user has not set them.
func InjectBuildEnv(existing []string, ramBudget, swapBudget int64) []string {
	seen := map[string]struct{}{}
	out := append([]string{}, existing...)
	for _, e := range existing {
		key, _, _ := strings.Cut(e, "=")
		seen[strings.ToUpper(strings.TrimSpace(key))] = struct{}{}
	}
	add := func(k, v string) {
		if _, ok := seen[strings.ToUpper(k)]; ok {
			return
		}
		out = append(out, k+"="+v)
	}
	totalMB := (ramBudget + swapBudget) / (1024 * 1024)
	heapMB := (totalMB * 3) / 4
	if heapMB < 512 {
		heapMB = 512
	}
	add("NODE_OPTIONS", fmt.Sprintf("--max-old-space-size=%d", heapMB))
	add("NEXT_TELEMETRY_DISABLED", "1")
	add("CI", "true")
	if hostMemGB() < 4 {
		add("UV_THREADPOOL_SIZE", "2")
	}
	return out
}

func hostMemGB() int64 {
	b, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 8
	}
	for _, line := range strings.Split(string(b), "\n") {
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
	return 8
}
