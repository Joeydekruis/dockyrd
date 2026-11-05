package deploy

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/build"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
)

func (rt *Runtime) acquireBuildBudget(ctx context.Context, deploymentID, serviceID string, expectedPeak int64) (build.Budget, int64, error) {
	if rt.Resources == nil || !rt.Resources.Available() {
		return build.Budget{}, 0, nil
	}
	waitStart := time.Now()
	var waitedMs int64
	for {
		snap, err := rt.Resources.Snapshot(ctx)
		if err != nil {
			return build.Budget{}, waitedMs, err
		}
		b := build.ComputeBudget(snap, build.BudgetParams{})
		dec := build.DecideAdmission(b, expectedPeak)
		if dec.Fail {
			_ = rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
				Status: "failed",
				Error:  sql.NullString{String: dec.Message, Valid: true},
				ID:     deploymentID,
			})
			return b, waitedMs, errors.New(dec.Message)
		}
		if dec.Allowed {
			if err := build.ApplyBuildLimits(ctx, rt.Resources, b); err != nil && rt.Log != nil {
				rt.Log.Warn("set build limits", "err", err)
			}
			return b, waitedMs, nil
		}
		_ = rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
			Status: "waiting",
			Error:  sql.NullString{String: dec.Message, Valid: true},
			ID:     deploymentID,
		})
		if rt.Events != nil {
			_ = rt.Events.Record(ctx, eventlog.RecordInput{
				Level:        "info",
				Category:     "build",
				Title:        "build.waiting",
				Detail:       dec.Message,
				ServiceID:    serviceID,
				DeploymentID: deploymentID,
			})
		}
		waitSlice := time.Now()
		select {
		case <-ctx.Done():
			waitedMs += time.Since(waitSlice).Milliseconds()
			return b, waitedMs, ctx.Err()
		case <-time.After(10 * time.Second):
		}
		waitedMs += time.Since(waitSlice).Milliseconds()
		if time.Since(waitStart) >= build.DefaultWaitTimeout() {
			dec.Allowed = true
			if err := build.ApplyBuildLimits(ctx, rt.Resources, b); err != nil && rt.Log != nil {
				rt.Log.Warn("set build limits after wait", "err", err)
			}
			return b, waitedMs, nil
		}
	}
}
