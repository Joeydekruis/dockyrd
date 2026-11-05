package watch

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/backup"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/notify"
)

// RunBackupsOverdue emits backup.overdue when policies miss successful runs.
func RunBackupsOverdue(ctx context.Context, q *sqlc.Queries, n *notify.Notifier, log *slog.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Minute
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			tickBackupsOverdue(ctx, q, n, log)
		}
	}
}

func tickBackupsOverdue(ctx context.Context, q *sqlc.Queries, n *notify.Notifier, log *slog.Logger) {
	if n == nil || q == nil {
		return
	}
	policies, err := q.ListEnabledDatabaseBackupPolicies(ctx)
	if err != nil {
		return
	}
	now := time.Now()
	for _, p := range policies {
		last, err := q.GetLatestSuccessfulBackupRun(ctx, p.ServiceID)
		var lastOK time.Time
		if err == nil && last.Valid {
			lastOK = time.Unix(last.Int64, 0)
		}
		base := lastOK
		if base.IsZero() {
			base = now.Add(-30 * 24 * time.Hour)
		}
		nextDue := backup.NextRunAfterLegacy(base, p.IntervalKind)
		grace := nextDue.Add(30 * time.Minute)
		dedup := "backup-overdue:" + p.ServiceID
		if now.After(grace) {
			detail := fmt.Sprintf("Expected interval %s, last successful %s.", p.IntervalKind, formatLast(lastOK))
			_ = n.Emit(ctx, notify.Event{
				Type: notify.TypeBackupOverdue, Level: notify.LevelWarning, State: notify.StateFiring,
				DedupKey: dedup, Title: "No recent backup · " + p.ServiceID, Detail: detail,
				ServiceID: p.ServiceID,
			})
		} else {
			_ = n.Emit(ctx, notify.Event{
				Type: notify.TypeBackupOverdue, Level: notify.LevelWarning, State: notify.StateResolved,
				DedupKey: dedup, Title: "Backup schedule OK", ServiceID: p.ServiceID,
			})
		}
	}
	_ = log
}

func formatLast(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.UTC().Format(time.RFC3339)
}
