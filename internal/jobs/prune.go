package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// PrunePayload targets one app service slug for image retention.
type PrunePayload struct {
	ServiceSlug string `json:"service_slug"`
}

const imageKeep = 5

// EnqueueImagePrune schedules image retention for one service.
func EnqueueImagePrune(ctx context.Context, q *sqlc.Queries, serviceSlug string) error {
	payload, _ := json.Marshal(PrunePayload{ServiceSlug: serviceSlug})
	now := time.Now().Unix()
	return q.EnqueueJob(ctx, sqlc.EnqueueJobParams{
		ID:        ulid.Make().String(),
		Kind:      sql.NullString{String: KindPrune, Valid: true},
		Payload:   payload,
		RunAfter:  sql.NullInt64{Int64: now, Valid: true},
		CreatedAt: sql.NullInt64{Int64: now, Valid: true},
	})
}

// RunImagePrune executes retention for one slug when Docker is available.
func RunImagePrune(ctx context.Context, c *docker.Client, slug string) error {
	if c == nil {
		return nil
	}
	return c.PruneServiceImages(ctx, slug, imageKeep)
}
