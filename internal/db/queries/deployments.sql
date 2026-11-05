-- name: CreateDeployment :exec
INSERT INTO deployments (
  id,
  service_id,
  status,
  trigger,
  triggered_by,
  image_ref,
  queued_at
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetDeploymentByID :one
SELECT
  id,
  service_id,
  status,
  trigger,
  triggered_by,
  commit_sha,
  commit_message,
  commit_author,
  branch,
  image_ref,
  release_path,
  error,
  build_log_path,
  mem_peak_bytes,
  swap_peak_bytes,
  cpu_seconds,
  waited_ms,
  frozen_ms,
  budget_bytes,
  builder,
  failure_kind,
  queued_at,
  started_at,
  finished_at
FROM deployments
WHERE id = ?
LIMIT 1;

-- name: ListDeploymentsForService :many
SELECT
  id,
  service_id,
  status,
  trigger,
  triggered_by,
  image_ref,
  release_path,
  commit_sha,
  commit_message,
  commit_author,
  branch,
  build_log_path,
  error,
  mem_peak_bytes,
  swap_peak_bytes,
  cpu_seconds,
  waited_ms,
  frozen_ms,
  budget_bytes,
  builder,
  failure_kind,
  queued_at,
  started_at,
  finished_at
FROM deployments
WHERE service_id = ?
ORDER BY queued_at DESC;

-- name: UpdateDeploymentStatus :exec
UPDATE deployments
SET
  status = ?,
  error = ?,
  started_at = COALESCE(?, started_at),
  finished_at = COALESCE(?, finished_at)
WHERE id = ?;

-- name: SetDeploymentImageRef :exec
UPDATE deployments
SET image_ref = ?
WHERE id = ?;

-- name: SetDeploymentBuildLogPath :exec
UPDATE deployments
SET build_log_path = ?
WHERE id = ?;

-- name: SetDeploymentReleasePath :exec
UPDATE deployments
SET release_path = ?
WHERE id = ?;

-- name: SetDeploymentGitMeta :exec
UPDATE deployments
SET
  commit_sha = ?,
  commit_message = ?,
  commit_author = ?,
  branch = ?
WHERE id = ?;

-- name: SetDeploymentBuildMetrics :exec
UPDATE deployments
SET
  mem_peak_bytes = ?,
  swap_peak_bytes = ?,
  waited_ms = ?,
  frozen_ms = ?,
  budget_bytes = ?,
  builder = ?,
  failure_kind = ?
WHERE id = ?;

-- name: SetDeploymentFailureKind :exec
UPDATE deployments
SET failure_kind = ?
WHERE id = ?;

-- name: SupersedeLiveDeployments :exec
UPDATE deployments
SET status = 'superseded', finished_at = ?
WHERE service_id = ? AND status = 'live' AND id != ?;

-- name: MarkInterruptedDeployments :exec
UPDATE deployments
SET
  status = 'failed',
  error = 'interrupted',
  finished_at = ?
WHERE status IN ('building', 'starting', 'healthchecking');

-- name: MarkInterruptedDeploymentsReturning :many
UPDATE deployments
SET
  status = 'failed',
  error = 'interrupted',
  finished_at = ?
WHERE status IN ('building', 'starting', 'healthchecking')
RETURNING id, service_id;
