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

-- name: SetServiceLiveDeployment :exec
UPDATE services
SET live_deployment_id = ?, updated_at = ?
WHERE id = ?;

-- name: ListServicesByProject :many
SELECT
  s.id,
  s.environment_id,
  s.kind,
  s.name,
  s.slug,
  s.desired_state,
  s.config,
  s.live_deployment_id,
  s.created_at,
  s.updated_at
FROM services s
JOIN environments e ON e.id = s.environment_id
WHERE e.project_id = ?
ORDER BY s.created_at DESC;

-- name: ListDeployableServices :many
SELECT id, kind, config, slug
FROM services
WHERE kind IN ('app', 'static_site');
