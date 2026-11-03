package server

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/deploy"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/events"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
	"github.com/Joeydekruis/dockyrd/internal/reconcile"
)

func startBackground(ctx context.Context, cfg config.Config, q *sqlc.Queries, hub *events.Hub) {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	var dockerCli *docker.Client
	cli, err := docker.NewClient()
	if err == nil {
		if err := cli.Ping(ctx); err == nil {
			dockerCli = cli
		} else {
			_ = cli.Close()
		}
	}

	deployRT := &deploy.Runtime{Cfg: cfg, Q: q, Docker: dockerCli, Log: log}
	worker := &jobs.Worker{Q: q, Deploy: deployRT, Log: log}
	rec := &reconcile.Loop{Cfg: cfg, Q: q, Docker: dockerCli, Log: log}

	go worker.RunLoop(ctx, 2*time.Second)
	go rec.Run(ctx, 30*time.Second)

	_ = hub // reserved for deployment SSE fan-out from handlers
}
