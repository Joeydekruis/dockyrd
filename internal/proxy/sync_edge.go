package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// SyncEdgeFromDB loads Caddy from SQLite, recreating dy-proxy when sync fails or :443 speaks plain HTTP.
func SyncEdgeFromDB(ctx context.Context, cfg config.Config, q *sqlc.Queries, dock *docker.Client, log *slog.Logger) error {
	if q == nil {
		return nil
	}
	if dock == nil {
		return SyncFromDB(ctx, q, cfg, nil)
	}
	if err := dock.EnsureProxy(ctx, cfg); err != nil {
		return err
	}
	last := trySync(ctx, cfg, q, dock, 3)
	if last == nil && !PlainHTTPOn443() {
		return nil
	}
	if last != nil && log != nil {
		log.Warn("proxy sync failed; recreating dy-proxy", "err", last)
	} else if PlainHTTPOn443() && log != nil {
		log.Warn("port 443 speaks plain HTTP; recreating dy-proxy and reloading routes")
	}
	if err := dock.RecreateProxy(ctx, cfg); err != nil {
		if last != nil {
			return fmt.Errorf("%w; recreate dy-proxy: %v", last, err)
		}
		return err
	}
	last = trySync(ctx, cfg, q, dock, 5)
	if last != nil {
		return last
	}
	if PlainHTTPOn443() {
		return fmt.Errorf("caddy still serves plain HTTP on :443 after reload (check panel logs and caddy load errors)")
	}
	return nil
}

func trySync(ctx context.Context, cfg config.Config, q *sqlc.Queries, dock *docker.Client, attempts int) error {
	var last error
	for i := 0; i < attempts; i++ {
		err := SyncFromDB(ctx, q, cfg, dock)
		if err == nil {
			return nil
		}
		last = err
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(i+1) * time.Second):
		}
	}
	return last
}
