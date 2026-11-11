package deploy

import (
	"context"
	"errors"
)

// ErrDeploymentCancelled indicates the deployment was cancelled before completion.
var ErrDeploymentCancelled = errors.New("deployment cancelled")

func (rt *Runtime) ensureNotCancelled(ctx context.Context, deploymentID string) error {
	dep, err := rt.Q.GetDeploymentByID(ctx, deploymentID)
	if err != nil {
		return err
	}
	if dep.Status == "cancelled" {
		return ErrDeploymentCancelled
	}
	return nil
}
