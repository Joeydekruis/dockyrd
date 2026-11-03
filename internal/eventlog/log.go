package eventlog

import (
	"context"
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/events"
)

const ResMinute = 60

// Logger persists platform events and publishes SSE updates.
type Logger struct {
	Q   *sqlc.Queries
	Hub *events.Hub
}

// RecordInput describes a new event row.
type RecordInput struct {
	Level        string
	Category     string
	Title        string
	Detail       string
	ProjectID    string
	ServiceID    string
	DeploymentID string
}

// Record inserts an event and notifies SSE subscribers.
func (l *Logger) Record(ctx context.Context, in RecordInput) error {
	if l == nil || l.Q == nil {
		return nil
	}
	now := time.Now().Unix()
	id := ulid.Make().String()
	if err := l.Q.InsertEvent(ctx, sqlc.InsertEventParams{
		ID:             id,
		Level:          sql.NullString{String: in.Level, Valid: in.Level != ""},
		Category:       sql.NullString{String: in.Category, Valid: in.Category != ""},
		Title:          sql.NullString{String: in.Title, Valid: in.Title != ""},
		Detail:         sql.NullString{String: in.Detail, Valid: in.Detail != ""},
		ProjectID:      sql.NullString{String: in.ProjectID, Valid: in.ProjectID != ""},
		ServiceID:      sql.NullString{String: in.ServiceID, Valid: in.ServiceID != ""},
		DeploymentID:   sql.NullString{String: in.DeploymentID, Valid: in.DeploymentID != ""},
		AcknowledgedAt: sql.NullInt64{},
		CreatedAt:      sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		return err
	}
	if l.Hub != nil {
		l.Hub.Publish("events", map[string]any{
			"id":       id,
			"level":    in.Level,
			"category": in.Category,
			"title":    in.Title,
		})
	}
	return nil
}
