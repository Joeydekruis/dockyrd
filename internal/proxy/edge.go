package proxy

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

var edgeRepairGate struct {
	sync.Mutex
	last time.Time
}

const autoRepairCooldown = 8 * time.Minute

// ReloadEdge ensures dy-proxy, optionally restarts it, and loads routes from the DB.
func ReloadEdge(ctx context.Context, cfg config.Config, q *sqlc.Queries, dock *docker.Client, restart bool, log *slog.Logger) error {
	if dock == nil || q == nil {
		return nil
	}
	if err := dock.EnsureProxy(ctx, cfg); err != nil {
		return err
	}
	if restart {
		if err := dock.RestartProxy(ctx, cfg); err != nil && log != nil {
			log.Warn("restart dy-proxy", "err", err)
		}
	}
	return SyncFromDB(ctx, q, cfg, dock)
}

// EdgeNeedsRepair returns true when any hostname fails TLS or ACME HTTP probes.
func EdgeNeedsRepair(hosts []string) bool {
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if ACMEChallengeBlocked(h) || !TLSHealthy(h) {
			return true
		}
	}
	return false
}

// TryAutoRepair reloads the edge when unhealthy (rate-limited). forced skips cooldown.
func TryAutoRepair(ctx context.Context, cfg config.Config, q *sqlc.Queries, dock *docker.Client, hosts []string, forced bool, log *slog.Logger) error {
	if dock == nil || q == nil || len(hosts) == 0 {
		return nil
	}
	if !EdgeNeedsRepair(hosts) {
		return nil
	}
	if !forced {
		edgeRepairGate.Lock()
		if time.Since(edgeRepairGate.last) < autoRepairCooldown {
			edgeRepairGate.Unlock()
			return nil
		}
		edgeRepairGate.last = time.Now()
		edgeRepairGate.Unlock()
	} else {
		edgeRepairGate.Lock()
		edgeRepairGate.last = time.Now()
		edgeRepairGate.Unlock()
	}
	if log != nil {
		log.Info("edge repair: restarting dy-proxy and syncing routes")
	}
	return ReloadEdge(ctx, cfg, q, dock, true, log)
}
