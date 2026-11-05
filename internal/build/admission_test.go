package build

import (
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/resources"
)

func TestComputeBudget(t *testing.T) {
	snap := resources.Snapshot{
		MemTotal:      1024 * 1024 * 1024,
		AppsWS:        200 * 1024 * 1024,
		SystemWS:      192 * 1024 * 1024,
		SwapFree:      2 * 1024 * 1024 * 1024,
		KernelReserve: 256 * 1024 * 1024,
	}
	b := ComputeBudget(snap, BudgetParams{})
	if b.RAMBytes <= 0 {
		t.Fatalf("expected positive ram budget, got %d", b.RAMBytes)
	}
	if b.SwapBytes <= 0 {
		t.Fatalf("expected swap budget")
	}
	limits := BuildLimitsFromBudget(b)
	if limits.MemoryHigh >= limits.MemoryMax {
		t.Fatalf("MemoryHigh should be below MemoryMax")
	}
}

func TestDecideAdmission_waitAndFail(t *testing.T) {
	b := Budget{RAMBytes: 900 * 1024 * 1024, SwapBytes: 100 * 1024 * 1024}
	dec := DecideAdmission(b, 1600*1024*1024)
	if !dec.Wait || dec.Allowed {
		t.Fatalf("expected wait")
	}
	low := Budget{RAMBytes: 400 * 1024 * 1024, SwapBytes: 0}
	dec = DecideAdmission(low, 768*1024*1024)
	if !dec.Fail {
		t.Fatalf("expected fail below 512MB")
	}
}
