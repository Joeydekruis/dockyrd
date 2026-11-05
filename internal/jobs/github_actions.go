package jobs

import (
	"context"
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// RecordGitHubActionsPush creates a waiting deployment without enqueueing a server build.
func RecordGitHubActionsPush(ctx context.Context, q *sqlc.Queries, serviceID string, meta GitMeta) (string, error) {
	now := time.Now().Unix()
	deploymentID := ulid.Make().String()
	if err := q.CreateDeployment(ctx, sqlc.CreateDeploymentParams{
		ID:       deploymentID,
		ServiceID: serviceID,
		Status:   "waiting",
		Trigger:  "github_push",
		QueuedAt: sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		return "", err
	}
	_ = q.SetDeploymentGitMeta(ctx, sqlc.SetDeploymentGitMetaParams{
		CommitSha:     nullStr(meta.CommitSHA),
		CommitMessage: nullStr(meta.CommitMessage),
		CommitAuthor:  nullStr(meta.CommitAuthor),
		Branch:        nullStr(meta.Branch),
		ID:            deploymentID,
	})
	_ = q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status: "waiting",
		Error:  sql.NullString{String: "Building on GitHub Actions…", Valid: true},
		ID:     deploymentID,
	})
	return deploymentID, nil
}

// EnqueueHookDeploy schedules pull+deploy for a pre-built image from CI.
func EnqueueHookDeploy(ctx context.Context, q *sqlc.Queries, serviceID, imageRef string, meta GitMeta) (string, error) {
	now := time.Now().Unix()
	deploymentID := ulid.Make().String()
	if err := q.CreateDeployment(ctx, sqlc.CreateDeploymentParams{
		ID:       deploymentID,
		ServiceID: serviceID,
		Status:   "queued",
		Trigger:  "deploy_hook",
		ImageRef: sql.NullString{String: imageRef, Valid: true},
		QueuedAt: sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		return "", err
	}
	_ = q.SetDeploymentGitMeta(ctx, sqlc.SetDeploymentGitMetaParams{
		CommitSha:     nullStr(meta.CommitSHA),
		CommitMessage: nullStr(meta.CommitMessage),
		CommitAuthor:  nullStr(meta.CommitAuthor),
		Branch:        nullStr(meta.Branch),
		ID:            deploymentID,
	})
	if err := EnqueueDeploy(ctx, q, deploymentID); err != nil {
		return "", err
	}
	return deploymentID, nil
}
