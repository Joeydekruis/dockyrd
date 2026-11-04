-- name: GetBackupTargetForServer :one
SELECT
  id,
  server_id,
  provider,
  endpoint,
  bucket,
  prefix,
  region,
  access_key_enc,
  secret_key_enc,
  created_at,
  updated_at
FROM backup_targets
WHERE server_id = ?
LIMIT 1;

-- name: UpsertBackupTarget :exec
INSERT INTO backup_targets (
  id,
  server_id,
  provider,
  endpoint,
  bucket,
  prefix,
  region,
  access_key_enc,
  secret_key_enc,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(server_id) DO UPDATE SET
  provider = excluded.provider,
  endpoint = excluded.endpoint,
  bucket = excluded.bucket,
  prefix = excluded.prefix,
  region = excluded.region,
  access_key_enc = excluded.access_key_enc,
  secret_key_enc = excluded.secret_key_enc,
  updated_at = excluded.updated_at;

-- name: GetDatabaseBackupPolicy :one
SELECT
  id,
  service_id,
  enabled,
  interval_kind,
  retention_days,
  local_enabled,
  remote_enabled,
  last_run_at,
  next_run_at,
  created_at,
  updated_at
FROM database_backup_policies
WHERE service_id = ?
LIMIT 1;

-- name: UpsertDatabaseBackupPolicy :exec
INSERT INTO database_backup_policies (
  id,
  service_id,
  enabled,
  interval_kind,
  retention_days,
  local_enabled,
  remote_enabled,
  last_run_at,
  next_run_at,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(service_id) DO UPDATE SET
  enabled = excluded.enabled,
  interval_kind = excluded.interval_kind,
  retention_days = excluded.retention_days,
  local_enabled = excluded.local_enabled,
  remote_enabled = excluded.remote_enabled,
  updated_at = excluded.updated_at;

-- name: ListDatabaseBackupRuns :many
SELECT
  id,
  service_id,
  policy_id,
  status,
  destination,
  object_key,
  size_bytes,
  started_at,
  finished_at,
  error
FROM database_backup_runs
WHERE service_id = ?
ORDER BY started_at DESC
LIMIT ?;

-- name: GetDatabaseBackupRun :one
SELECT
  id,
  service_id,
  policy_id,
  status,
  destination,
  object_key,
  size_bytes,
  started_at,
  finished_at,
  error
FROM database_backup_runs
WHERE id = ?
LIMIT 1;

-- name: InsertDatabaseBackupRun :exec
INSERT INTO database_backup_runs (
  id,
  service_id,
  policy_id,
  status,
  destination,
  object_key,
  size_bytes,
  started_at,
  finished_at,
  error
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateDatabaseBackupRun :exec
UPDATE database_backup_runs
SET
  status = ?,
  object_key = ?,
  size_bytes = ?,
  finished_at = ?,
  error = ?
WHERE id = ?;

-- name: ListDueDatabaseBackupPolicies :many
SELECT
  p.id,
  p.service_id,
  p.enabled,
  p.interval_kind,
  p.retention_days,
  p.local_enabled,
  p.remote_enabled,
  p.last_run_at,
  p.next_run_at,
  p.created_at,
  p.updated_at,
  s.kind AS service_kind,
  s.slug AS service_slug,
  e.project_id,
  pr.slug AS project_slug
FROM database_backup_policies p
JOIN services s ON s.id = p.service_id
JOIN environments e ON e.id = s.environment_id
JOIN projects pr ON pr.id = e.project_id
WHERE p.enabled = 1
  AND (p.next_run_at IS NULL OR p.next_run_at <= ?);

-- name: UpdateDatabaseBackupPolicySchedule :exec
UPDATE database_backup_policies
SET last_run_at = ?, next_run_at = ?, updated_at = ?
WHERE id = ?;

-- name: ListDatabaseBackupRunsForPrune :many
SELECT id, object_key, started_at
FROM database_backup_runs
WHERE service_id = ?
  AND status = 'succeeded'
  AND destination = ?
  AND started_at IS NOT NULL
  AND started_at < ?
ORDER BY started_at ASC;
