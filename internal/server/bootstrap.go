package server

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// EnsureLocalServer creates the MVP local server row once.
func EnsureLocalServer(ctx context.Context, q *sqlc.Queries) (sqlc.Server, error) {
	row, err := q.GetLocalServer(ctx)
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return sqlc.Server{}, err
	}
	now := time.Now().Unix()
	id := ulid.Make().String()
	if err := q.CreateServer(ctx, sqlc.CreateServerParams{
		ID:        id,
		Name:      sql.NullString{String: "Local server", Valid: true},
		Kind:      sql.NullString{String: "local", Valid: true},
		CreatedAt: sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		return sqlc.Server{}, err
	}
	return q.GetServerByID(ctx, id)
}
