package reconcile

import (
	"context"
	"log/slog"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/storage"
)

// SyncGarage ensures dy-garage is running when installed.
func SyncGarage(ctx context.Context, cfg config.Config, q *sqlc.Queries, dc *docker.Client, log *slog.Logger) {
	if dc == nil {
		return
	}
	ok, err := storage.IsGarageInstalled(ctx, q)
	if err != nil || !ok {
		return
	}
	if err := storage.EnsureGarageRunning(ctx, cfg, dc); err != nil && log != nil {
		log.Warn("garage reconcile", "err", err)
	}
}
