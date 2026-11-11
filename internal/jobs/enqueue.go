package jobs

import (
	"context"
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// GitMeta optional commit metadata stored on a deployment.
type GitMeta struct {
	CommitSHA     string
	CommitMessage string
	CommitAuthor  string
	Branch        string
}

// EnqueueServiceDeploy creates a queued deployment and schedules the worker job.
func EnqueueServiceDeploy(ctx context.Context, q *sqlc.Queries, serviceID, trigger, userID string, meta GitMeta) (string, error) {
	if err := EnsureDeploymentSlot(ctx, q); err != nil {
		return "", err
	}
	now := time.Now().Unix()
	deploymentID := ulid.Make().String()
	triggeredBy := sql.NullString{}
	if userID != "" {
		triggeredBy = sql.NullString{String: userID, Valid: true}
	}
	if err := q.CreateDeployment(ctx, sqlc.CreateDeploymentParams{
		ID:          deploymentID,
		ServiceID:   serviceID,
		Status:      "queued",
		Trigger:     trigger,
		TriggeredBy: triggeredBy,
		ImageRef:    sql.NullString{},
		QueuedAt:    sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		return "", err
	}
	if meta.CommitSHA != "" || meta.Branch != "" {
		_ = q.SetDeploymentGitMeta(ctx, sqlc.SetDeploymentGitMetaParams{
			CommitSha:     nullStr(meta.CommitSHA),
			CommitMessage: nullStr(meta.CommitMessage),
			CommitAuthor:  nullStr(meta.CommitAuthor),
			Branch:        nullStr(meta.Branch),
			ID:            deploymentID,
		})
	}
	if err := EnqueueDeploy(ctx, q, deploymentID); err != nil {
		return "", err
	}
	return deploymentID, nil
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
