package server

import (
	"context"
	"log/slog"
	"os"

	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

func (d apiDeps) syncEdgeProxy(ctx context.Context) {
	if d.q == nil || d.docker == nil {
		return
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := d.docker.EnsureProxy(ctx, d.cfg); err != nil {
		log.Warn("ensure dy-proxy", "err", err)
		return
	}
	if err := proxy.SyncFromDB(ctx, d.q, d.cfg); err != nil {
		log.Warn("proxy sync", "err", err)
	}
}
