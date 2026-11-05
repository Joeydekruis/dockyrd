package server

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

// RepairStaleLoopbackPlatformDomains rewrites sslip.io hostnames that still encode 127.0.0.1.
func RepairStaleLoopbackPlatformDomains(ctx context.Context, cfg config.Config, q *sqlc.Queries) {
	if q == nil {
		return
	}
	effective := cfg.EffectivePublicIPv4()
	if effective == "127.0.0.1" {
		return
	}
	rows, err := q.ListStaleLoopbackPlatformDomains(ctx)
	if err != nil || len(rows) == 0 {
		return
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	for _, row := range rows {
		svc, err := q.GetServiceByID(ctx, row.ServiceID)
		if err != nil {
			continue
		}
		host, err := proxy.PlatformHostname(svc.Slug, effective)
		if err != nil {
			continue
		}
		if strings.EqualFold(host, row.Hostname) {
			continue
		}
		if err := q.UpdateDomainHostname(ctx, sqlc.UpdateDomainHostnameParams{
			Hostname: host,
			ID:       row.ID,
		}); err != nil {
			log.Warn("repair platform domain hostname", "id", row.ID, "err", err)
			continue
		}
		log.Info("repaired loopback platform domain", "from", row.Hostname, "to", host)
	}
}

func (d apiDeps) repairStaleLoopbackPlatformDomains(ctx context.Context) {
	RepairStaleLoopbackPlatformDomains(ctx, d.cfg, d.q)
}
