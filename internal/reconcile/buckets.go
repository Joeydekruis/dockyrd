package reconcile

import (
	"context"
	"log/slog"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// SyncBucketNetworks attaches dy-garage to project networks that have buckets.
func SyncBucketNetworks(ctx context.Context, q *sqlc.Queries, dc *docker.Client, log *slog.Logger) {
	if dc == nil {
		return
	}
	rows, err := q.ListServicesByKind(ctx, "bucket")
	if err != nil {
		if log != nil {
			log.Warn("bucket network list", "err", err)
		}
		return
	}
	seen := map[string]struct{}{}
	for _, row := range rows {
		if row.ProjectSlug == "" {
			continue
		}
		if _, ok := seen[row.ProjectSlug]; ok {
			continue
		}
		seen[row.ProjectSlug] = struct{}{}
		if err := dc.EnsureGarageOnProjectNetwork(ctx, row.ProjectSlug); err != nil && log != nil {
			log.Debug("garage project network", "project", row.ProjectSlug, "err", err)
		}
	}
}
