package watch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
	"github.com/Joeydekruis/dockyrd/internal/metrics"
	"github.com/Joeydekruis/dockyrd/internal/notify"
)

const diskResolvePct = 80

type diskSample struct {
	mount   string
	usedPct float64
	free    int64
	total   int64
}

// diskAlertDecision describes whether to emit and at which level/state.
type diskAlertDecision struct {
	emit  bool
	level notify.Level
	state notify.State
}

// evalDiskAlert maps usage percentage to firing/resolver/no-op per briefing thresholds.
func evalDiskAlert(usedPct, warn, crit, resolve float64) diskAlertDecision {
	switch {
	case usedPct >= crit:
		return diskAlertDecision{emit: true, level: notify.LevelError, state: notify.StateFiring}
	case usedPct >= warn:
		return diskAlertDecision{emit: true, level: notify.LevelWarning, state: notify.StateFiring}
	case usedPct < resolve:
		return diskAlertDecision{emit: true, level: notify.LevelWarning, state: notify.StateResolved}
	default:
		return diskAlertDecision{}
	}
}

// RunDisk emits disk.high alerts from metric_host (data dir) and Docker root when on another mount.
func RunDisk(ctx context.Context, cfg config.Config, q *sqlc.Queries, n *notify.Notifier, log *slog.Logger, interval time.Duration, dockerCli *docker.Client) {
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
			tickDisk(ctx, cfg, q, n, log, dockerCli)
		}
	}
}

func tickDisk(ctx context.Context, cfg config.Config, q *sqlc.Queries, n *notify.Notifier, log *slog.Logger, dockerCli *docker.Client) {
	if n == nil || q == nil {
		return
	}
	settings := loadNotifySettings(ctx, q)
	warn := float64(settings.DiskWarnPct)
	crit := float64(settings.DiskCritPct)
	resolve := float64(diskResolvePct)

	for _, sample := range collectDiskSamples(ctx, cfg, q, dockerCli) {
		emitDiskSample(ctx, n, log, sample, warn, crit, resolve)
	}
}

func collectDiskSamples(ctx context.Context, cfg config.Config, q *sqlc.Queries, dockerCli *docker.Client) []diskSample {
	var out []diskSample
	if s := diskSampleFromMetrics(ctx, q, cfg.DataDir); s != nil {
		out = append(out, *s)
	}
	if dockerCli == nil {
		return out
	}
	root, err := dockerCli.DockerRootDir(ctx)
	if err != nil || root == "" {
		return out
	}
	same, err := sameFilesystem(cfg.DataDir, root)
	if err != nil || same {
		return out
	}
	if s := diskSampleFromPath(root); s != nil {
		out = append(out, *s)
	}
	return out
}

func diskSampleFromMetrics(ctx context.Context, q *sqlc.Queries, mount string) *diskSample {
	row, err := q.GetLatestMetricHost(ctx, sqlNullInt64(eventlog.ResMinute))
	if err != nil {
		return nil
	}
	if !row.DiskTotal.Valid || row.DiskTotal.Int64 <= 0 || !row.DiskUsed.Valid {
		return nil
	}
	usedPct := float64(row.DiskUsed.Int64) * 100 / float64(row.DiskTotal.Int64)
	free := row.DiskTotal.Int64 - row.DiskUsed.Int64
	return &diskSample{
		mount:   mount,
		usedPct: usedPct,
		free:    free,
		total:   row.DiskTotal.Int64,
	}
}

func diskSampleFromPath(mount string) *diskSample {
	sample, err := metrics.SampleHost(mount)
	if err != nil || sample.DiskTotal <= 0 {
		return nil
	}
	usedPct := float64(sample.DiskUsed) * 100 / float64(sample.DiskTotal)
	free := sample.DiskTotal - sample.DiskUsed
	return &diskSample{
		mount:   mount,
		usedPct: usedPct,
		free:    free,
		total:   sample.DiskTotal,
	}
}

func emitDiskSample(ctx context.Context, n *notify.Notifier, log *slog.Logger, sample diskSample, warn, crit, resolve float64) {
	dec := evalDiskAlert(sample.usedPct, warn, crit, resolve)
	if !dec.emit {
		if log != nil {
			log.Debug("disk watch", "mount", sample.mount, "pct", sample.usedPct)
		}
		return
	}
	dedup := "disk:" + sample.mount
	detail := fmt.Sprintf("%s is %.0f%% full (%s free of %s).",
		sample.mount, sample.usedPct, bytesGiB(sample.free), bytesGiB(sample.total))

	threshold := warn
	title := fmt.Sprintf("Disk usage is high (%.0f%%)", sample.usedPct)
	if dec.state == notify.StateResolved {
		title = "Disk usage back to normal"
		detail = fmt.Sprintf("%s now at %.0f%%.", sample.mount, sample.usedPct)
	} else if dec.level == notify.LevelError {
		threshold = crit
		title = fmt.Sprintf("Disk almost full (%.0f%%)", sample.usedPct)
		if log != nil {
			log.Error("disk almost full", "mount", sample.mount, "used_percent", sample.usedPct)
		}
	}

	_ = n.Emit(ctx, notify.Event{
		Type:     notify.TypeDiskHigh,
		Level:    dec.level,
		State:    dec.state,
		DedupKey: dedup,
		Title:    title,
		Detail:   detail,
		Href:     "/server",
		Data: map[string]any{
			"mount":        sample.mount,
			"used_percent": sample.usedPct,
			"threshold":    threshold,
		},
	})
}

func bytesGiB(b int64) string {
	if b < 0 {
		b = 0
	}
	return fmt.Sprintf("%.1f GiB", float64(b)/(1024*1024*1024))
}
