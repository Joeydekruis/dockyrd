package resources

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	dbus "github.com/godbus/dbus/v5"
	systemdDbus "github.com/coreos/go-systemd/v22/dbus"
)

// Linux reads cgroups and sets slice limits over systemd D-Bus.
type Linux struct{}

func newLinux() (*Linux, error) {
	if !cgroupV2Unified() {
		return nil, errors.New("cgroup v2 required")
	}
	return &Linux{}, nil
}

func (l *Linux) Available() bool { return true }

func (l *Linux) Snapshot(ctx context.Context) (Snapshot, error) {
	_ = ctx
	memTotal, err := readMemTotal()
	if err != nil {
		return Snapshot{}, err
	}
	swapFree, _ := readSwapFree()
	appsWS, err := sliceWorkingSet(SliceApps)
	if err != nil {
		appsWS = 0
	}
	systemWS, err := sliceWorkingSet(SliceSystem)
	if err != nil {
		systemWS = 192 * 1024 * 1024
	}
	const systemFloor = 192 * 1024 * 1024
	if systemWS < systemFloor {
		systemWS = systemFloor
	}
	kernel := kernelReserve(memTotal)
	return Snapshot{
		MemTotal:      memTotal,
		AppsWS:        appsWS,
		SystemWS:      systemWS,
		SwapFree:      swapFree,
		KernelReserve: kernel,
	}, nil
}

func kernelReserve(memTotal int64) int64 {
	const minReserve = 256 * 1024 * 1024
	pct := (memTotal * 8) / 100
	if pct < minReserve {
		return minReserve
	}
	return pct
}

func (l *Linux) SetBuildLimits(ctx context.Context, limits BuildLimits) error {
	conn, err := systemdDbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return fallbackSetBuildLimits(limits)
	}
	defer conn.Close()

	props := []systemdDbus.Property{
		{Name: "MemoryHigh", Value: dbus.MakeVariant(uint64(limits.MemoryHigh))},
		{Name: "MemoryMax", Value: dbus.MakeVariant(uint64(limits.MemoryMax))},
		{Name: "MemorySwapMax", Value: dbus.MakeVariant(uint64(limits.SwapMax))},
	}
	return conn.SetUnitPropertiesContext(ctx, SliceBuild, true, props...)
}

func fallbackSetBuildLimits(limits BuildLimits) error {
	base := sliceCgroupPath(SliceBuild)
	if err := os.WriteFile(filepath.Join(base, "memory.high"), []byte(fmt.Sprintf("%d", limits.MemoryHigh)), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(base, "memory.max"), []byte(fmt.Sprintf("%d", limits.MemoryMax)), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(base, "memory.swap.max"), []byte(fmt.Sprintf("%d", limits.SwapMax)), 0o644)
}

func (l *Linux) FreezeBuild(ctx context.Context, frozen bool) error {
	_ = ctx
	return writeCgroupFreeze(SliceBuild, frozen)
}

func (l *Linux) BuildUsage(ctx context.Context) (Usage, error) {
	_ = ctx
	base := sliceCgroupPath(SliceBuild)
	mem, _ := readCgroupInt64(filepath.Join(base, "memory.current"))
	swap, _ := readCgroupInt64(filepath.Join(base, "memory.swap.current"))
	oom, _ := readBuildOOMKills()
	return Usage{MemCurrent: mem, SwapCurrent: swap, OOMKillCount: oom}, nil
}
