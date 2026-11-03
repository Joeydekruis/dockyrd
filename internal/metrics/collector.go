package metrics

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
)

// RunHostCollector samples host metrics every interval until ctx is cancelled.
func RunHostCollector(ctx context.Context, cfg config.Config, q *sqlc.Queries, log *slog.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			sample, err := SampleHost(cfg.DataDir)
			if err != nil {
				if log != nil {
					log.Debug("host metrics sample", "err", err)
				}
				continue
			}
			ts := MinuteTS(now)
			if err := q.InsertMetricHost(ctx, sqlc.InsertMetricHostParams{
				Ts:        sql.NullInt64{Int64: ts, Valid: true},
				Res:       sql.NullInt64{Int64: eventlog.ResMinute, Valid: true},
				Cpu:       sql.NullFloat64{Float64: sample.CPU, Valid: true},
				MemUsed:   sql.NullInt64{Int64: sample.MemUsed, Valid: true},
				MemTotal:  sql.NullInt64{Int64: sample.MemTotal, Valid: true},
				DiskUsed:  sql.NullInt64{Int64: sample.DiskUsed, Valid: true},
				DiskTotal: sql.NullInt64{Int64: sample.DiskTotal, Valid: true},
				NetRx:     sql.NullInt64{Int64: sample.NetRx, Valid: true},
				NetTx:     sql.NullInt64{Int64: sample.NetTx, Valid: true},
				Load1:     sql.NullFloat64{Float64: sample.Load1, Valid: true},
			}); err != nil && log != nil {
				log.Debug("host metrics insert", "err", err)
			}
		}
	}
}
