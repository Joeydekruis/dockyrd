-- name: InsertEvent :exec
INSERT INTO events (
  id,
  level,
  category,
  title,
  detail,
  project_id,
  service_id,
  deployment_id,
  acknowledged_at,
  created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListRecentEvents :many
SELECT
  id,
  level,
  category,
  title,
  detail,
  project_id,
  service_id,
  deployment_id,
  acknowledged_at,
  created_at
FROM events
ORDER BY created_at DESC
LIMIT ?;

-- name: AckEvent :exec
UPDATE events
SET acknowledged_at = ?
WHERE id = ?;

-- name: GetEventByID :one
SELECT
  id,
  level,
  category,
  title,
  detail,
  project_id,
  service_id,
  deployment_id,
  acknowledged_at,
  created_at
FROM events
WHERE id = ?
LIMIT 1;
