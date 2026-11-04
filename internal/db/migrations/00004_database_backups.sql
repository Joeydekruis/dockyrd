-- +goose Up
CREATE TABLE backup_targets (
  id TEXT PRIMARY KEY,
  server_id TEXT NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
  provider TEXT NOT NULL DEFAULT 's3' CHECK (provider IN ('s3')),
  endpoint TEXT NOT NULL DEFAULT '',
  bucket TEXT NOT NULL,
  prefix TEXT NOT NULL DEFAULT '',
  region TEXT NOT NULL DEFAULT '',
  access_key_enc BLOB,
  secret_key_enc BLOB,
  created_at INTEGER,
  updated_at INTEGER,
  UNIQUE(server_id)
);

CREATE TABLE database_backup_policies (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  enabled INTEGER NOT NULL DEFAULT 1,
  interval_kind TEXT NOT NULL DEFAULT 'daily' CHECK (interval_kind IN ('hourly', 'daily', 'weekly')),
  retention_days INTEGER NOT NULL DEFAULT 30 CHECK (retention_days > 0),
  local_enabled INTEGER NOT NULL DEFAULT 1,
  remote_enabled INTEGER NOT NULL DEFAULT 0,
  last_run_at INTEGER,
  next_run_at INTEGER,
  created_at INTEGER,
  updated_at INTEGER,
  UNIQUE(service_id)
);

CREATE TABLE database_backup_runs (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  policy_id TEXT REFERENCES database_backup_policies(id) ON DELETE SET NULL,
  status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'succeeded', 'failed')),
  destination TEXT NOT NULL CHECK (destination IN ('local', 'remote')),
  object_key TEXT NOT NULL DEFAULT '',
  size_bytes INTEGER NOT NULL DEFAULT 0,
  started_at INTEGER,
  finished_at INTEGER,
  error TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_database_backup_runs_service ON database_backup_runs(service_id, started_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_database_backup_runs_service;
DROP TABLE IF EXISTS database_backup_runs;
DROP TABLE IF EXISTS database_backup_policies;
DROP TABLE IF EXISTS backup_targets;
