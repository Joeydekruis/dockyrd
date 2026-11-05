package resources

import (
	"context"
	"runtime"
)

// BuildLimits are runtime cgroup limits for dy-build.slice (bytes).
type BuildLimits struct {
	MemoryHigh int64
	MemoryMax  int64
	SwapMax    int64
}

// Snapshot is a point-in-time view of host memory and slice working sets.
type Snapshot struct {
	MemTotal   int64
	AppsWS     int64
	SystemWS   int64
	SwapFree   int64
	KernelReserve int64
}

// Usage is current build slice consumption.
type Usage struct {
	MemCurrent   int64
	SwapCurrent  int64
	OOMKillCount uint64
}

// Resources abstracts cgroup/systemd interactions (Linux prod) or no-ops (dev/tests).
type Resources interface {
	Available() bool
	Snapshot(ctx context.Context) (Snapshot, error)
	SetBuildLimits(ctx context.Context, l BuildLimits) error
	FreezeBuild(ctx context.Context, frozen bool) error
	BuildUsage(ctx context.Context) (Usage, error)
}

// New returns Linux implementation when isolation is enabled on linux/amd64|arm64, else Fake.
func New(isolationEnabled bool) Resources {
	if !isolationEnabled || runtime.GOOS != "linux" {
		return &Fake{}
	}
	if l, err := newLinux(); err == nil {
		return l
	}
	return &Fake{}
}
