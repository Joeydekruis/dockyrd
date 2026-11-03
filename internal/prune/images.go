package prune

import (
	"context"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

const imageKeep = 5

// AllAppServiceImages prunes dockyard/<slug> tags for every app service.
func AllAppServiceImages(ctx context.Context, q *sqlc.Queries, c *docker.Client, keep int) {
	if q == nil || c == nil {
		return
	}
	if keep <= 0 {
		keep = imageKeep
	}
	rows, err := q.ListDeployableServices(ctx)
	if err != nil {
		return
	}
	for _, row := range rows {
		if row.Kind != "app" {
			continue
		}
		_ = c.PruneServiceImages(ctx, row.Slug, keep)
	}
}
