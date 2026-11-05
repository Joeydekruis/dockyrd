package build

import (
	"context"
	"fmt"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/resources"
)

// BudgetParams tune the build RAM/swap budget formula.
type BudgetParams struct {
	HeadroomPct float64 // default 0.15 of apps working set
}

// Budget is computed RAM and swap a build may use right now.
type Budget struct {
	RAMBytes  int64
	SwapBytes int64
	Snapshot  resources.Snapshot
}

// ComputeBudget derives build slice limits from a host snapshot.
func ComputeBudget(snap resources.Snapshot, p BudgetParams) Budget {
	if p.HeadroomPct <= 0 {
		p.HeadroomPct = 0.15
	}
	headroom := int64(float64(snap.AppsWS) * p.HeadroomPct)
	const minHeadroom = 128 * 1024 * 1024
	if headroom < minHeadroom {
		headroom = minHeadroom
	}
	ram := snap.MemTotal - snap.KernelReserve - snap.SystemWS - snap.AppsWS - headroom
	if ram < 0 {
		ram = 0
	}
	swap := snap.SwapFree - 256*1024*1024
	const maxSwapBudget = 2 * 1024 * 1024 * 1024
	if swap < 0 {
		swap = 0
	}
	if swap > maxSwapBudget {
		swap = maxSwapBudget
	}
	return Budget{RAMBytes: ram, SwapBytes: swap, Snapshot: snap}
}

// BuildLimitsFromBudget maps budget to cgroup properties (MemoryHigh = 90% RAM).
func BuildLimitsFromBudget(b Budget) resources.BuildLimits {
	high := (b.RAMBytes * 9) / 10
	if high < 0 {
		high = 0
	}
	return resources.BuildLimits{
		MemoryHigh: high,
		MemoryMax:  b.RAMBytes,
		SwapMax:    b.SwapBytes,
	}
}

// ExpectedPeakHeuristic returns initial peak estimate when no history exists.
func ExpectedPeakHeuristic(builder string, isNextJS bool) int64 {
	switch {
	case isNextJS:
		return int64(1.5 * 1024 * 1024 * 1024)
	case builder == "dockerfile", builder == "railpack":
		return 1024 * 1024 * 1024
	default:
		return 768 * 1024 * 1024
	}
}

// AdmissionDecision is whether a build can start now.
type AdmissionDecision struct {
	Allowed       bool
	Wait          bool
	Fail          bool
	Message       string
	Budget        Budget
	ExpectedPeak  int64
	WaitTimeout   time.Duration
}

const (
	minRAMToBuild      = 512 * 1024 * 1024
	defaultWaitTimeout = 15 * time.Minute
)

// DefaultWaitTimeout is how long admission waits before starting with available budget.
func DefaultWaitTimeout() time.Duration { return defaultWaitTimeout }

// DecideAdmission compares expected peak to ram+swap budget.
func DecideAdmission(b Budget, expectedPeak int64) AdmissionDecision {
	total := b.RAMBytes + b.SwapBytes
	dec := AdmissionDecision{
		Budget:       b,
		ExpectedPeak: expectedPeak,
		WaitTimeout:  defaultWaitTimeout,
	}
	if b.RAMBytes < minRAMToBuild {
		dec.Fail = true
		dec.Message = fmt.Sprintf(
			"Not enough memory to build safely (%d MB available). Add swap, use a remote builder, or build on GitHub Actions.",
			b.RAMBytes/(1024*1024),
		)
		return dec
	}
	if total >= expectedPeak {
		dec.Allowed = true
		return dec
	}
	dec.Wait = true
	dec.Message = fmt.Sprintf(
		"Waiting for resources: needs ~%.1f GB, %.1f GB available",
		float64(expectedPeak)/(1024*1024*1024),
		float64(total)/(1024*1024*1024),
	)
	return dec
}

// WaitForAdmission blocks until the build may start or ctx ends.
func WaitForAdmission(
	ctx context.Context,
	res resources.Resources,
	expectedPeak int64,
	pollEvery time.Duration,
	waitStarted time.Time,
) (AdmissionDecision, error) {
	if pollEvery <= 0 {
		pollEvery = 10 * time.Second
	}
	for {
		snap, err := res.Snapshot(ctx)
		if err != nil {
			return AdmissionDecision{}, err
		}
		b := ComputeBudget(snap, BudgetParams{})
		dec := DecideAdmission(b, expectedPeak)
		if dec.Allowed {
			return dec, nil
		}
		if dec.Fail {
			return dec, nil
		}
		if time.Since(waitStarted) >= defaultWaitTimeout && b.RAMBytes >= minRAMToBuild {
			dec.Allowed = true
			dec.Wait = false
			dec.Message = "Starting build with available memory budget after wait timeout"
			return dec, nil
		}
		select {
		case <-ctx.Done():
			return dec, ctx.Err()
		case <-time.After(pollEvery):
		}
	}
}

// ApplyBuildLimits sets dy-build.slice limits from a budget.
func ApplyBuildLimits(ctx context.Context, res resources.Resources, b Budget) error {
	return res.SetBuildLimits(ctx, BuildLimitsFromBudget(b))
}
