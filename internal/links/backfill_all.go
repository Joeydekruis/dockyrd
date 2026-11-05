package links

import (
	"context"
	"log/slog"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// BackfillAllProjects scans every project once at startup (idempotent).
func BackfillAllProjects(ctx context.Context, q *sqlc.Queries, secretKeyPath string, log *slog.Logger) {
	if q == nil {
		return
	}
	key, err := crypto.LoadMasterKey(secretKeyPath)
	if err != nil {
		return
	}
	projects, err := q.ListProjects(ctx)
	if err != nil {
		return
	}
	for _, p := range projects {
		created, err := BackfillProject(ctx, q, key, p.ID)
		if err != nil && log != nil {
			log.Warn("link backfill", "project", p.Slug, "err", err)
			continue
		}
		if created > 0 && log != nil {
			log.Info("link backfill", "project", p.Slug, "created", created)
		}
	}
}
