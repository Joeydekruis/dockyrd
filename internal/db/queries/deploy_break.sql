-- name: UpsertDeploymentBreakScore :exec
INSERT INTO deployment_break_scores (id, deployment_id, user_id, score, updated_at)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(deployment_id, user_id) DO UPDATE SET
  score = excluded.score,
  updated_at = excluded.updated_at
WHERE excluded.score > deployment_break_scores.score;

-- name: GetDeploymentBreakScoreForUser :one
SELECT score, updated_at
FROM deployment_break_scores
WHERE deployment_id = ? AND user_id = ?
LIMIT 1;

-- name: GetDeploymentBreakBest :one
SELECT
  dbs.score,
  dbs.updated_at,
  u.name AS user_name,
  u.id AS user_id
FROM deployment_break_scores dbs
JOIN users u ON u.id = dbs.user_id
WHERE dbs.deployment_id = ?
ORDER BY dbs.score DESC, dbs.updated_at ASC
LIMIT 1;

-- name: GetDeploymentBreakAllTimeBest :one
SELECT
  dbs.score,
  dbs.updated_at,
  dbs.deployment_id,
  u.name AS user_name,
  u.id AS user_id,
  s.name AS service_name,
  s.slug AS service_slug,
  s.kind AS service_kind,
  p.slug AS project_slug
FROM deployment_break_scores dbs
JOIN users u ON u.id = dbs.user_id
JOIN deployments d ON d.id = dbs.deployment_id
JOIN services s ON s.id = d.service_id
JOIN environments e ON e.id = s.environment_id
JOIN projects p ON p.id = e.project_id
ORDER BY dbs.score DESC, dbs.updated_at ASC
LIMIT 1;
