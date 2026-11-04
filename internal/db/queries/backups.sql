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
