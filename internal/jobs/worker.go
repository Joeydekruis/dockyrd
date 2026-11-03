package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/deploy"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

const (
	KindDeploy = "deploy"
	KindPrune  = "prune"
	lockSeconds = 120
)

// DeployPayload is stored on deploy jobs.
type DeployPayload struct {
	DeploymentID string `json:"deployment_id"`
}

// Worker processes the SQLite job queue.
type Worker struct {
	Q       *sqlc.Queries
	Deploy  *deploy.Runtime
	Docker  *docker.Client
	Log     *slog.Logger
}

// EnqueueDeploy schedules a deploy job.
func EnqueueDeploy(ctx context.Context, q *sqlc.Queries, deploymentID string) error {
	payload, _ := json.Marshal(DeployPayload{DeploymentID: deploymentID})
	now := time.Now().Unix()
	return q.EnqueueJob(ctx, sqlc.EnqueueJobParams{
		ID:        ulid.Make().String(),
		Kind:      sql.NullString{String: KindDeploy, Valid: true},
		Payload:   payload,
		RunAfter:  sql.NullInt64{Int64: now, Valid: true},
		CreatedAt: sql.NullInt64{Int64: now, Valid: true},
	})
}

func payloadBytes(v any) ([]byte, error) {
	switch t := v.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	default:
		return json.Marshal(v)
	}
}

// ProcessOnce claims and runs one pending job.
func (w *Worker) ProcessOnce(ctx context.Context) (bool, error) {
	now := time.Now().Unix()
	row, err := w.Q.ClaimNextJob(ctx, sql.NullInt64{Int64: now, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	lockUntil := now + lockSeconds
	if err := w.Q.LockJob(ctx, sqlc.LockJobParams{
		LockedUntil: sql.NullInt64{Int64: lockUntil, Valid: true},
		ID:          row.ID,
	}); err != nil {
		return false, err
	}

	var runErr error
	switch {
	case row.Kind.Valid && row.Kind.String == KindDeploy:
		var payload DeployPayload
		raw, err := payloadBytes(row.Payload)
		if err != nil {
			runErr = err
		} else if err := json.Unmarshal(raw, &payload); err != nil {
			runErr = err
		} else if w.Deploy != nil {
			runErr = w.Deploy.RunDeployment(ctx, payload.DeploymentID)
		} else {
			runErr = errors.New("deploy runtime not configured")
		}
	case row.Kind.Valid && row.Kind.String == KindPrune:
		var payload PrunePayload
		raw, err := payloadBytes(row.Payload)
		if err != nil {
			runErr = err
		} else if err := json.Unmarshal(raw, &payload); err != nil {
			runErr = err
		} else {
			runErr = RunImagePrune(ctx, w.Docker, payload.ServiceSlug)
		}
	default:
		runErr = errors.New("unknown job kind")
	}

	if runErr != nil {
		_ = w.Q.FailJob(ctx, sqlc.FailJobParams{
			LastError: sql.NullString{String: runErr.Error(), Valid: true},
			ID:        row.ID,
		})
		if w.Log != nil {
			w.Log.Error("job failed", "kind", row.Kind, "err", runErr)
		}
		return true, runErr
	}
	_ = w.Q.CompleteJob(ctx, row.ID)
	return true, nil
}

// RunLoop polls until ctx is cancelled.
func (w *Worker) RunLoop(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			for {
				done, err := w.ProcessOnce(ctx)
				if err != nil || !done {
					break
				}
			}
		}
	}
}
