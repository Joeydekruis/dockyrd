package server

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/deploy"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
	"github.com/Joeydekruis/dockyrd/internal/events"
	"github.com/Joeydekruis/dockyrd/internal/inbox"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
	"github.com/Joeydekruis/dockyrd/internal/links"
	"github.com/Joeydekruis/dockyrd/internal/metrics"
	"github.com/Joeydekruis/dockyrd/internal/notify"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
	"github.com/Joeydekruis/dockyrd/internal/reconcile"
	"github.com/Joeydekruis/dockyrd/internal/resources"
	"github.com/Joeydekruis/dockyrd/internal/updater"
	"github.com/Joeydekruis/dockyrd/internal/watch"
)

func startBackground(ctx context.Context, cfg config.Config, db *sql.DB, q *sqlc.Queries, hub *events.Hub, notifier *notify.Notifier, dispatcher *notify.Dispatcher) {
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

	eventLog := &eventlog.Logger{Q: q, Hub: hub}
	inboxDispatch := &inbox.Dispatcher{Cfg: cfg, Q: q, Hub: hub, Log: log}
	deployRT := &deploy.Runtime{
		Cfg:       cfg,
		Q:         q,
		Docker:    dockerCli,
		Log:       log,
		Events:    eventLog,
		Inbox:     inboxDispatch,
		Notify:    notifier,
		Resources: resources.New(cfg.BuildIsolationEnabled()),
	}
	rec := &reconcile.Loop{
		Cfg:       cfg,
		Q:         q,
		Docker:    dockerCli,
		Log:       log,
		Resources: resources.New(cfg.BuildIsolationEnabled()),
		Notify:    notifier,
	}

	worker := &jobs.Worker{Cfg: cfg, Q: q, Deploy: deployRT, Docker: dockerCli, Log: log, Notify: notifier, Events: eventLog}

	go metrics.RunHostCollector(ctx, cfg, q, log, 10*time.Second)
	if dockerCli != nil {
		go metrics.RunContainerCollector(ctx, dockerCli, q, log, 15*time.Second)
	}

	if dockerCli != nil {
		if err := dockerCli.EnsureEdgeNetwork(ctx); err != nil && log != nil {
			log.Warn("ensure dy-edge", "err", err)
		}
		if err := dockerCli.EnsureBuildkit(ctx, cfg.DataDir); err != nil && log != nil {
			log.Warn("ensure dy-buildkit", "err", err)
		}
		if err := dockerCli.EnsureProxy(ctx, cfg); err != nil && log != nil {
			log.Warn("ensure dy-proxy", "err", err)
		} else {
			RepairStaleLoopbackPlatformDomains(ctx, cfg, q)
			if err := proxy.SyncFromDB(ctx, q, cfg, dockerCli); err != nil && log != nil {
				log.Warn("proxy sync", "err", err)
			}
		}
	}

	go worker.RunLoop(ctx, 2*time.Second)
	go rec.Run(ctx, 30*time.Second)
	if dispatcher != nil {
		go dispatcher.Run(ctx, 2*time.Second)
	}
	go watch.RunDNS(ctx, cfg, q, log, 2*time.Minute)
	go watch.RunCerts(ctx, cfg, q, notifier, log, 6*time.Hour)
	if notifier != nil {
		go watch.RunDisk(ctx, cfg, q, notifier, log, 60*time.Second, dockerCli)
		go watch.RunBackupsOverdue(ctx, q, notifier, log, 10*time.Minute)
	}
	go watch.RunHeartbeat(ctx, q, log, 5*time.Minute)

	go links.BackfillAllProjects(ctx, q, cfg.SecretKeyPath(), log)

	updater.StartAutoRefresh(ctx, updater.AutoRefreshInterval, func() string {
		return BuildVersion()
	}, func() updater.Options {
		return updater.Options{Repo: os.Getenv("DOCKYARD_GITHUB_REPO")}
	})

	_ = hub
	_ = db
}
