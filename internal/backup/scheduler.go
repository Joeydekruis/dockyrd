package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// EnqueueDuePolicies schedules backup jobs for policies whose next_run_at has passed.
func EnqueueDuePolicies(ctx context.Context, q *sqlc.Queries, now int64) error {
	if q == nil {
		return nil
	}
	rows, err := q.ListDueDatabaseBackupPolicies(ctx, sql.NullInt64{Int64: now, Valid: true})
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.LocalEnabled != 1 && row.RemoteEnabled != 1 {
			continue
		}
		payload, _ := json.Marshal(JobPayload{
			ServiceID: row.ServiceID,
			PolicyID:  row.ID,
		})
		if err := q.EnqueueJob(ctx, sqlc.EnqueueJobParams{
			ID:        ulid.Make().String(),
			Kind:      sql.NullString{String: "database_backup", Valid: true},
			Payload:   payload,
			RunAfter:  sql.NullInt64{Int64: now, Valid: true},
			CreatedAt: sql.NullInt64{Int64: now, Valid: true},
		}); err != nil {
			return err
		}
		next := NextRunAfter(time.Unix(now, 0), row.IntervalKind).Unix()
		if err := q.UpdateDatabaseBackupPolicySchedule(ctx, sqlc.UpdateDatabaseBackupPolicyScheduleParams{
			LastRunAt: sql.NullInt64{Int64: row.LastRunAt.Int64, Valid: row.LastRunAt.Valid},
			NextRunAt: sql.NullInt64{Int64: next, Valid: true},
			UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
			ID:        row.ID,
		}); err != nil {
			return err
		}
	}
	return nil
}

// EnqueueManualBackup schedules an immediate backup for one service.
func EnqueueManualBackup(ctx context.Context, q *sqlc.Queries, serviceID, policyID string) error {
	payload, _ := json.Marshal(JobPayload{ServiceID: serviceID, PolicyID: policyID})
	now := time.Now().Unix()
	return q.EnqueueJob(ctx, sqlc.EnqueueJobParams{
		ID:        ulid.Make().String(),
		Kind:      sql.NullString{String: "database_backup", Valid: true},
		Payload:   payload,
		RunAfter:  sql.NullInt64{Int64: now, Valid: true},
		CreatedAt: sql.NullInt64{Int64: now, Valid: true},
	})
}
