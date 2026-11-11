-- name: ListAppServices :many
SELECT
  s.id,
  s.name,
  s.slug,
  s.desired_state,
  p.id AS project_id,
  p.name AS project_name,
  p.slug AS project_slug,
  (
    SELECT d.hostname
    FROM domains d
    WHERE d.service_id = s.id
    ORDER BY d.is_primary DESC, d.hostname ASC
    LIMIT 1
  ) AS primary_hostname
FROM services s
JOIN environments e ON e.id = s.environment_id
JOIN projects p ON p.id = e.project_id
WHERE s.kind = 'app'
ORDER BY p.name ASC, s.name ASC;

-- name: GetServiceProjectSlug :one
SELECT p.slug
FROM services s
JOIN environments e ON e.id = s.environment_id
JOIN projects p ON p.id = e.project_id
WHERE s.id = ?
LIMIT 1;
