-- name: ListProjects :many
SELECT
  id,
  server_id,
  name,
  slug,
  description,
  created_at,
  updated_at
FROM projects
ORDER BY created_at DESC;

-- name: GetProjectByID :one
SELECT
  id,
  server_id,
  name,
  slug,
  description,
  created_at,
  updated_at
FROM projects
WHERE id = ?
LIMIT 1;

-- name: GetProjectBySlug :one
SELECT
  id,
  server_id,
  name,
  slug,
  description,
  created_at,
  updated_at
FROM projects
WHERE slug = ?
LIMIT 1;

-- name: CreateProject :exec
INSERT INTO projects (
  id,
  server_id,
  name,
  slug,
  description,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateProject :exec
UPDATE projects
SET
  name = ?,
  description = ?,
  updated_at = ?
WHERE id = ?;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = ?;

-- name: AddProjectMember :exec
INSERT INTO project_members (project_id, user_id, role)
VALUES (?, ?, ?);
