package watch

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

// RunEdgeRepair probes TLS/ACME health and reloads dy-proxy when needed.
func RunEdgeRepair(ctx context.Context, cfg config.Config, q *sqlc.Queries, dock *docker.Client, log *slog.Logger, interval time.Duration) {
	if dock == nil || q == nil {
		return
	}
	if interval <= 0 {
		interval = 3 * time.Minute
	}
	TickEdgeRepair(ctx, cfg, q, dock, log, false)
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			TickEdgeRepair(ctx, cfg, q, dock, log, false)
		}
	}
}

// TickEdgeRepair runs one edge health pass (forced ignores auto-repair cooldown).
func TickEdgeRepair(ctx context.Context, cfg config.Config, q *sqlc.Queries, dock *docker.Client, log *slog.Logger, forced bool) {
	hosts := EdgeProbeHostnames(ctx, cfg, q)
	if len(hosts) == 0 {
		return
	}
	if err := proxy.TryAutoRepair(ctx, cfg, q, dock, hosts, forced, log); err != nil && log != nil {
		log.Warn("edge repair", "err", err)
	}
}

// EdgeProbeHostnames lists hostnames used for local TLS/ACME health checks.
func EdgeProbeHostnames(ctx context.Context, cfg config.Config, q *sqlc.Queries) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(h string) {
		h = strings.TrimSpace(h)
		if h == "" {
			return
		}
		h = strings.TrimSuffix(strings.ToLower(h), ".")
		if _, ok := seen[h]; ok {
			return
		}
		seen[h] = struct{}{}
		out = append(out, h)
	}

	panel, err := instance.LoadPanel(ctx, q)
	if err == nil {
		pub := instance.EffectivePublicURL(cfg, panel)
		if u, err := url.Parse(pub); err == nil {
			add(u.Hostname())
		}
		if w := strings.TrimSpace(panel.WildcardProbeHost); w != "" {
			add(w)
		}
	}

	rows, err := q.ListDomainsForCertWatch(ctx)
	if err == nil {
		for _, row := range rows {
			if row.DesiredState != "running" {
				continue
			}
			add(row.Hostname)
		}
	}
	return out
}
