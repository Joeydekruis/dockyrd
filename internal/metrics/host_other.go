//go:build !linux

package metrics

import "runtime"

// SampleHost returns a best-effort sample on non-Linux dev hosts.
func SampleHost(_ string) (HostSample, error) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return HostSample{
		MemUsed:  int64(ms.Alloc),
		MemTotal: int64(ms.Sys),
		CPU:      0,
		Load1:    0,
	}, nil
}
