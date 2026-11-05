package deploy

import (
	"context"
	"database/sql"

	"github.com/Joeydekruis/dockyrd/internal/build"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func (rt *Runtime) recordBuildMetrics(ctx context.Context, deploymentID string, budget build.Budget, builder string, sampler *build.PeakSampler) {
	if rt.Q == nil {
		return
	}
	memPeak, swapPeak := sampler.Stop()
	var memNull, swapNull, budgetNull sql.NullInt64
	if memPeak > 0 {
		memNull = sql.NullInt64{Int64: memPeak, Valid: true}
	}
	if swapPeak > 0 {
		swapNull = sql.NullInt64{Int64: swapPeak, Valid: true}
	}
	totalBudget := budget.RAMBytes + budget.SwapBytes
	if totalBudget > 0 {
		budgetNull = sql.NullInt64{Int64: totalBudget, Valid: true}
	}
	if builder == "" {
		builder = "local"
	}
	_ = rt.Q.SetDeploymentBuildMetrics(ctx, sqlc.SetDeploymentBuildMetricsParams{
		MemPeakBytes:  memNull,
		SwapPeakBytes: swapNull,
		WaitedMs:      0,
		FrozenMs:      0,
		BudgetBytes:   budgetNull,
		Builder:       builder,
		FailureKind:   sql.NullString{},
		ID:            deploymentID,
	})
}

func (rt *Runtime) failWithKind(ctx context.Context, deploymentID, kind string, err error) error {
	msg := build.FailureMessage(kind, err.Error())
	_ = rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status: "failed",
		Error:  sql.NullString{String: msg, Valid: true},
		ID:     deploymentID,
	})
	if kind != "" {
		_ = rt.Q.SetDeploymentFailureKind(ctx, sqlc.SetDeploymentFailureKindParams{
			FailureKind: sql.NullString{String: kind, Valid: true},
			ID:          deploymentID,
		})
	}
	return err
}
