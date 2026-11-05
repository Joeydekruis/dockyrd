package deploy

import (
	"context"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func (rt *Runtime) skipProjectNetwork(ctx context.Context, projectID string) (bool, error) {
	mode, err := rt.Q.GetProjectLinksMode(ctx, projectID)
	if err != nil {
		return false, err
	}
	return mode == "strict", nil
}

// SkipProjectNetworkForProject is used by reconciler when starting databases.
func SkipProjectNetworkForProject(ctx context.Context, q *sqlc.Queries, projectID string) (bool, error) {
	rt := Runtime{Q: q}
	return rt.skipProjectNetwork(ctx, projectID)
}
