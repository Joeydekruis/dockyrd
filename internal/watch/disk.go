package watch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
	"github.com/Joeydekruis/dockyrd/internal/notify"
)

// RunDisk emits disk.high alerts from metric_host samples.
func RunDisk(ctx context.Context, cfg config.Config, q *sqlc.Queries, n *notify.Notifier, log *slog.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = 60 * time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			tickDisk(ctx, cfg, q, n, log)
		}
	}
}

func tickDisk(ctx context.Context, cfg config.Config, q *sqlc.Queries, n *notify.Notifier, log *slog.Logger) {
	if n == nil || q == nil {
		return
	}
	row, err := q.GetLatestMetricHost(ctx, sqlNullInt64(eventlog.ResMinute))
	if err != nil {
		return
	}
	if !row.DiskTotal.Valid || row.DiskTotal.Int64 <= 0 || !row.DiskUsed.Valid {
		return
	}
	usedPct := float64(row.DiskUsed.Int64) * 100 / float64(row.DiskTotal.Int64)
	settings := loadNotifySettings(ctx, q)
	warn := float64(settings.DiskWarnPct)
	crit := float64(settings.DiskCritPct)
	resolveBelow := warn - 5
	if resolveBelow < 1 {
		resolveBelow = 80
	}
	dedup := "disk:" + cfg.DataDir
	free := row.DiskTotal.Int64 - row.DiskUsed.Int64
	detail := fmt.Sprintf("%s is %.0f%% full (%s free of %s).",
		cfg.DataDir, usedPct, bytesGiB(free), bytesGiB(row.DiskTotal.Int64))

	switch {
	case usedPct >= crit:
		_ = n.Emit(ctx, notify.Event{
			Type: notify.TypeDiskHigh, Level: notify.LevelError, State: notify.StateFiring,
			DedupKey: dedup, Title: fmt.Sprintf("Disk almost full (%.0f%%)", usedPct), Detail: detail, Href: "/server",
			Data: map[string]any{"mount": cfg.DataDir, "used_percent": usedPct, "threshold": crit},
		})
	case usedPct >= warn:
		_ = n.Emit(ctx, notify.Event{
			Type: notify.TypeDiskHigh, Level: notify.LevelWarning, State: notify.StateFiring,
			DedupKey: dedup, Title: fmt.Sprintf("Disk usage is high (%.0f%%)", usedPct), Detail: detail, Href: "/server",
			Data: map[string]any{"mount": cfg.DataDir, "used_percent": usedPct, "threshold": warn},
		})
	case usedPct < resolveBelow:
		_ = n.Emit(ctx, notify.Event{
			Type: notify.TypeDiskHigh, Level: notify.LevelWarning, State: notify.StateResolved,
			DedupKey: dedup, Title: "Disk usage back to normal", Detail: fmt.Sprintf("Now at %.0f%%.", usedPct),
		})
	default:
		if log != nil {
			log.Debug("disk watch", "pct", usedPct)
		}
	}
}

func bytesGiB(b int64) string {
	if b < 0 {
		b = 0
	}
	return fmt.Sprintf("%.1f GiB", float64(b)/(1024*1024*1024))
}
