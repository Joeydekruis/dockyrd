package reconcile

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/notify"
	"github.com/Joeydekruis/dockyrd/internal/prune"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
	"github.com/Joeydekruis/dockyrd/internal/resources"
)

// Loop marks interrupted deploys and syncs proxy config periodically.
type Loop struct {
	Cfg               config.Config
	Q                 *sqlc.Queries
	Docker            *docker.Client
	Log               *slog.Logger
	Resources         resources.Resources
	Notify            *notify.Notifier
	tickCount         int
	thawed            bool
	interruptedMarked bool
}

// Tick runs one reconcile pass.
func (l *Loop) Tick(ctx context.Context) {
	if !l.thawed && l.Resources != nil {
		if err := l.Resources.FreezeBuild(ctx, false); err != nil && l.Log != nil {
			l.Log.Warn("thaw build slice", "err", err)
		}
		l.thawed = true
	}
	if !l.interruptedMarked {
		now := time.Now().Unix()
		rows, err := l.Q.MarkInterruptedDeploymentsReturning(ctx, sql.NullInt64{Int64: now, Valid: true})
		if err != nil && l.Log != nil {
			l.Log.Warn("mark interrupted deployments", "err", err)
		} else if l.Notify != nil {
			for _, row := range rows {
				l.Notify.EmitDeployFailed(ctx, row.ServiceID, row.ID, "Deploy interrupted by panel restart")
			}
		}
		l.interruptedMarked = true
	}
	if l.Docker != nil {
		if err := l.Docker.Ping(ctx); err != nil {
			return
		}
		SyncDatabaseServices(ctx, l.Cfg, l.Q, l.Docker, l.Log)
		SyncGarage(ctx, l.Cfg, l.Q, l.Docker, l.Log)
		SyncServiceLinks(ctx, l.Q, l.Docker, l.Log)
		SyncBucketNetworks(ctx, l.Q, l.Docker, l.Log)
	}
	if err := proxy.SyncFromDB(ctx, l.Q, l.Cfg); err != nil && l.Log != nil {
		l.Log.Debug("proxy sync", "err", err)
	}
	l.tickCount++
	if l.Docker != nil && l.tickCount%10 == 0 {
		prune.AllAppServiceImages(ctx, l.Q, l.Docker, 5)
	}
}

// Run starts the 30s reconcile ticker.
func (l *Loop) Run(ctx context.Context, interval time.Duration) {
	l.Tick(ctx)
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			l.Tick(ctx)
		}
	}
}
