package reconcile

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/prune"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

// Loop marks interrupted deploys and syncs proxy config periodically.
type Loop struct {
	Cfg       config.Config
	Q         *sqlc.Queries
	Docker    *docker.Client
	Log       *slog.Logger
	tickCount int
}

// Tick runs one reconcile pass.
func (l *Loop) Tick(ctx context.Context) {
	now := time.Now().Unix()
	if err := l.Q.MarkInterruptedDeployments(ctx, sql.NullInt64{Int64: now, Valid: true}); err != nil && l.Log != nil {
		l.Log.Warn("mark interrupted deployments", "err", err)
	}
	if l.Docker != nil {
		if err := l.Docker.Ping(ctx); err != nil {
			return
		}
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
