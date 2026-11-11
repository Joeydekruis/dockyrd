package jobs

import (
	"context"
	"errors"
	"fmt"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// ErrDeploymentSlotBusy is returned when another deployment is already in progress.
var ErrDeploymentSlotBusy = errors.New("another deployment is in progress")

// EnsureDeploymentSlot checks that no deployment is currently in progress.
func EnsureDeploymentSlot(ctx context.Context, q *sqlc.Queries) error {
	n, err := q.CountInProgressDeployments(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("%w", ErrDeploymentSlotBusy)
	}
	return nil
}
