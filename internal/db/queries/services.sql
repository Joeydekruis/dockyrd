-- name: GetEnvironmentByProjectAndName :one
SELECT id, project_id, name
FROM environments
WHERE project_id = ? AND name = ?
LIMIT 1;

-- name: CreateService :exec
INSERT INTO services (
  id,
  environment_id,
  kind,
  name,
  slug,
  desired_state,
  config,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetServiceByID :one
SELECT
  id,
  environment_id,
  kind,
  name,
  slug,
  desired_state,
  config,
  live_deployment_id,
  created_at,
  updated_at
FROM services
WHERE id = ?
LIMIT 1;
