package docker

import (
	"context"
	"encoding/json"
	"io"

	"github.com/docker/docker/api/types/container"
)

// ContainerMetricSample is one stats read for a managed service.
type ContainerMetricSample struct {
	ServiceID string
	CPUFraction float64
	MemBytes    int64
	MemLimit    int64
	NetRx       int64
	NetTx       int64
}

// SampleManagedStats reads one-shot Docker stats for running managed containers.
func (c *Client) SampleManagedStats(ctx context.Context) ([]ContainerMetricSample, error) {
	if c == nil || c.moby == nil {
		return nil, ErrUnavailable
	}
	managed, err := c.ListManaged(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ContainerMetricSample, 0, len(managed))
	for _, m := range managed {
		if m.State != "running" || m.Identity.ServiceID == "" {
			continue
		}
		sample, err := c.statsOnce(ctx, m.ID)
		if err != nil {
			continue
		}
		sample.ServiceID = m.Identity.ServiceID
		out = append(out, sample)
	}
	return out, nil
}

func (c *Client) statsOnce(ctx context.Context, containerID string) (ContainerMetricSample, error) {
	var out ContainerMetricSample
	resp, err := c.moby.ContainerStats(ctx, containerID, false)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}
	var v container.StatsResponse
	if err := json.Unmarshal(body, &v); err != nil {
		return out, err
	}
	cpuPct := cpuPercentFromStats(v)
	out.CPUFraction = cpuPct / 100
	if out.CPUFraction < 0 {
		out.CPUFraction = 0
	}
	if out.CPUFraction > 1 {
		out.CPUFraction = 1
	}
	out.MemBytes = int64(v.MemoryStats.Usage)
	out.MemLimit = int64(v.MemoryStats.Limit)
	for _, nw := range v.Networks {
		out.NetRx += int64(nw.RxBytes)
		out.NetTx += int64(nw.TxBytes)
	}
	return out, nil
}

func cpuPercentFromStats(v container.StatsResponse) float64 {
	var (
		cpuDelta  = float64(v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta = float64(v.CPUStats.SystemUsage - v.PreCPUStats.SystemUsage)
	)
	if cpuDelta <= 0 || systemDelta <= 0 {
		return 0
	}
	online := v.CPUStats.OnlineCPUs
	if online == 0 {
		online = uint32(len(v.CPUStats.CPUUsage.PercpuUsage))
	}
	if online == 0 {
		online = 1
	}
	return (cpuDelta / systemDelta) * float64(online) * 100
}
