package metrics

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
)

// RunContainerCollector samples Docker stats for managed services until ctx is cancelled.
func RunContainerCollector(ctx context.Context, cli *docker.Client, q *sqlc.Queries, log *slog.Logger, interval time.Duration) {
	if cli == nil || q == nil {
		return
	}
	if interval <= 0 {
		interval = 15 * time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			samples, err := cli.SampleManagedStats(ctx)
			if err != nil {
				if log != nil {
					log.Debug("container metrics sample", "err", err)
				}
				continue
			}
			ts := MinuteTS(now)
			for _, s := range samples {
				if s.ServiceID == "" {
					continue
				}
				if err := q.InsertMetricContainer(ctx, sqlc.InsertMetricContainerParams{
					Ts:        sql.NullInt64{Int64: ts, Valid: true},
					Res:       sql.NullInt64{Int64: eventlog.ResMinute, Valid: true},
					ServiceID: sql.NullString{String: s.ServiceID, Valid: true},
					Cpu:       sql.NullFloat64{Float64: s.CPUFraction, Valid: true},
					Mem:       sql.NullInt64{Int64: s.MemBytes, Valid: true},
					MemLimit:  sql.NullInt64{Int64: s.MemLimit, Valid: true},
					NetRx:     sql.NullInt64{Int64: s.NetRx, Valid: true},
					NetTx:     sql.NullInt64{Int64: s.NetTx, Valid: true},
					Restarts:  sql.NullInt64{Int64: 0, Valid: true},
				}); err != nil && log != nil {
					log.Debug("container metrics insert", "err", err, "service", s.ServiceID)
				}
			}
		}
	}
}
