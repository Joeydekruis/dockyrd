package hostops

import (
	"context"
	"database/sql"
	"log/slog"
)

// Reconcile reverts awaiting_confirm rows past confirm_deadline (startup safety pass).
func Reconcile(ctx context.Context, r *Runner, log *slog.Logger) (int, error) {
	if r == nil || r.Q == nil {
		return 0, nil
	}
	now := r.Now().Unix()
	rows, err := r.Q.ListHostActionsAwaitingConfirmExpired(ctx, sql.NullInt64{Int64: now, Valid: true})
	if err != nil {
		return 0, err
	}
	reverted := 0
	for _, row := range rows {
		err := r.Revert(ctx, row.ID, RevertMeta{Source: "startup_reconcile"})
		if err != nil {
			if log != nil {
				log.Error("hostops reconcile revert", "id", row.ID, "err", err)
			}
			continue
		}
		reverted++
	}
	if log != nil && reverted > 0 {
		log.Info("hostops reconcile", "reverted", reverted, "candidates", len(rows))
	}
	return reverted, nil
}
