package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
)

func TestCPUPercentFromStats(t *testing.T) {
	t.Parallel()
	v := container.StatsResponse{
		PreCPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{TotalUsage: 100},
			SystemUsage: 1000,
		},
		CPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{TotalUsage: 200},
			SystemUsage: 2000,
			OnlineCPUs: 2,
		},
	}
	got := cpuPercentFromStats(v)
	if got <= 0 {
		t.Fatalf("expected positive cpu %%, got %v", got)
	}
}
