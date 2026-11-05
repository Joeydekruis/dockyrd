package watch

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/netsafe"
)

// RunHeartbeat pings an external URL when configured (dead-man's switch for external uptime tools).
func RunHeartbeat(ctx context.Context, q *sqlc.Queries, log *slog.Logger, interval time.Duration) {
	if q == nil {
		return
	}
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	tick := func() {
		s := loadNotifySettings(ctx, q)
		url := strings.TrimSpace(s.HeartbeatURL)
		if url == "" {
			return
		}
		client := netsafe.HTTPClient(netsafe.ClientOptions{Timeout: 15 * time.Second})
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			if log != nil {
				log.Debug("heartbeat request", "err", err)
			}
			return
		}
		res, err := client.Do(req)
		if err != nil {
			if log != nil {
				log.Debug("heartbeat ping failed", "err", err)
			}
			return
		}
		_ = res.Body.Close()
	}
	tick()
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			tick()
		}
	}
}
