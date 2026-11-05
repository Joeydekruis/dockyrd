package reconcile

import (
	"context"
	"log/slog"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/serviceops"
)

// SyncServiceLinks ensures per-link Docker networks connect live consumer and provider containers.
func SyncServiceLinks(ctx context.Context, q *sqlc.Queries, dc *docker.Client, log *slog.Logger) {
	if dc == nil || q == nil {
		return
	}
	rows, err := q.ListServiceLinksForReconcile(ctx)
	if err != nil {
		if log != nil {
			log.Warn("list service links", "err", err)
		}
		return
	}
	for _, row := range rows {
		if err := syncOneLink(ctx, dc, row); err != nil && log != nil {
			log.Debug("sync service link", "link", row.ID, "err", err)
		}
	}
}

func syncOneLink(ctx context.Context, dc *docker.Client, row sqlc.ListServiceLinksForReconcileRow) error {
	if row.FromLiveDeploymentID.Valid && row.FromLiveDeploymentID.String != "" {
		consumerID, err := dc.FindLiveContainerForService(ctx, row.FromServiceID, row.FromLiveDeploymentID.String)
		if err == nil {
			if err := dc.ConnectLinkEndpoint(ctx, row.ID, consumerID, nil); err != nil {
				return err
			}
		}
	}

	switch row.ToKind {
	case "bucket":
		return dc.EnsureGarageOnLinkNetwork(ctx, row.ID)
	case "postgres", "redis":
		providerID, err := serviceops.FindDatabaseContainerID(ctx, dc, row.ToServiceID)
		if err != nil {
			return nil
		}
		return dc.ConnectLinkEndpoint(ctx, row.ID, providerID, []string{row.ToSlug})
	case "app":
		if !row.ToLiveDeploymentID.Valid || row.ToLiveDeploymentID.String == "" {
			return nil
		}
		providerID, err := dc.FindLiveContainerForService(ctx, row.ToServiceID, row.ToLiveDeploymentID.String)
		if err != nil {
			return nil
		}
		return dc.ConnectLinkEndpoint(ctx, row.ID, providerID, []string{row.ToSlug})
	default:
		return nil
	}
}
