package server

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/project"
)

// EnsureUnassignedProject returns the instance bucket used when no project_id is supplied.
func EnsureUnassignedProject(ctx context.Context, q *sqlc.Queries) (sqlc.Project, error) {
	row, err := q.GetProjectBySlug(ctx, project.UnassignedSlug)
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return sqlc.Project{}, err
	}
	srv, err := EnsureLocalServer(ctx, q)
	if err != nil {
		return sqlc.Project{}, err
	}
	now := time.Now().Unix()
	projectID := ulid.Make().String()
	if err := q.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:          projectID,
		ServerID:    srv.ID,
		Name:        project.UnassignedName,
		Slug:        project.UnassignedSlug,
		Description: sql.NullString{},
		CreatedAt:   sql.NullInt64{Int64: now, Valid: true},
		UpdatedAt:   sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		return sqlc.Project{}, err
	}
	envID := ulid.Make().String()
	if err := q.CreateEnvironment(ctx, sqlc.CreateEnvironmentParams{
		ID:        envID,
		ProjectID: projectID,
		Name:      "production",
	}); err != nil {
		return sqlc.Project{}, err
	}
	_ = q.UpdateProjectLinksMode(ctx, sqlc.UpdateProjectLinksModeParams{
		LinksMode: "open",
		UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
		ID:        projectID,
	})
	return q.GetProjectBySlug(ctx, project.UnassignedSlug)
}

func resolveCreateProjectID(ctx context.Context, q *sqlc.Queries, projectID string) (string, error) {
	if projectID != "" {
		return projectID, nil
	}
	proj, err := EnsureUnassignedProject(ctx, q)
	if err != nil {
		return "", err
	}
	return proj.ID, nil
}
