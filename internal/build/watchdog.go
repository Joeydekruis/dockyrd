package build

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/resources"
)

// WatchdogThresholds tune freeze/thaw (defaults from build resilience briefing).
type WatchdogThresholds struct {
	FreezeAppsSomePct   float64
	FreezeSystemFullPct float64
	ThawAppsSomePct     float64
	ThawSystemFullPct   float64
	ThawStableFor       time.Duration
	MaxFrozenTotal      time.Duration
}

func defaultWatchdogThresholds() WatchdogThresholds {
	return WatchdogThresholds{
		FreezeAppsSomePct:   20,
		FreezeSystemFullPct: 10,
		ThawAppsSomePct:     5,
		ThawSystemFullPct:   2,
		ThawStableFor:       30 * time.Second,
		MaxFrozenTotal:      10 * time.Minute,
	}
}

// WatchdogCallbacks optional hooks for events / metrics.
type WatchdogCallbacks struct {
	OnFrozen  func()
	OnResumed func()
	OnFailed  func(reason string)
}

// RunBuildWatchdog monitors PSI and freezes the build slice under pressure.
func RunBuildWatchdog(
	ctx context.Context,
	res resources.Resources,
	cb WatchdogCallbacks,
	th WatchdogThresholds,
) {
	if res == nil || !res.Available() {
		return
	}
	if th.FreezeAppsSomePct <= 0 {
		th = defaultWatchdogThresholds()
	}
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	frozen := false
	var frozenTotal time.Duration
	var thawSince time.Time
	var initialOOM uint64
	if u, err := res.BuildUsage(ctx); err == nil {
		initialOOM = u.OOMKillCount
	}

	fail := func(msg string) {
		_ = res.FreezeBuild(ctx, false)
		if cb.OnFailed != nil {
			cb.OnFailed(msg)
		}
	}

	for {
		select {
		case <-ctx.Done():
			if frozen {
				_ = res.FreezeBuild(context.Background(), false)
			}
			return
		case <-ticker.C:
		}

		if u, err := res.BuildUsage(ctx); err == nil && u.OOMKillCount > initialOOM {
			fail("Build ran out of memory under cgroup limits")
			return
		}

		psi := resources.ReadPSI()
		if !frozen {
			if psi.AppsSomeAvg10 > th.FreezeAppsSomePct || psi.SystemFullAvg10 > th.FreezeSystemFullPct {
				if err := res.FreezeBuild(ctx, true); err == nil {
					frozen = true
					if cb.OnFrozen != nil {
						cb.OnFrozen()
					}
				}
			}
			continue
		}

		frozenTotal += 2 * time.Second
		if frozenTotal >= th.MaxFrozenTotal {
			fail("Build paused too long: your apps need the memory")
			return
		}

		if psi.AppsSomeAvg10 < th.ThawAppsSomePct && psi.SystemFullAvg10 < th.ThawSystemFullPct {
			if thawSince.IsZero() {
				thawSince = time.Now()
			}
			if time.Since(thawSince) >= th.ThawStableFor {
				if err := res.FreezeBuild(ctx, false); err == nil {
					frozen = false
					thawSince = time.Time{}
					if cb.OnResumed != nil {
						cb.OnResumed()
					}
				}
			}
		} else {
			thawSince = time.Time{}
		}
	}
}

// FrozenMsTracker accumulates time spent frozen (for deployment metrics).
type FrozenMsTracker struct {
	ms atomic.Int64
}

func (t *FrozenMsTracker) AddFrozen(d time.Duration) {
	if t == nil {
		return
	}
	t.ms.Add(d.Milliseconds())
}

func (t *FrozenMsTracker) Milliseconds() int64 {
	if t == nil {
		return 0
	}
	return t.ms.Load()
}
