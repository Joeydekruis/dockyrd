-- +goose Up
ALTER TABLE database_backup_policies ADD COLUMN run_at_minute INTEGER NOT NULL DEFAULT 180;
ALTER TABLE database_backup_policies ADD COLUMN run_weekday INTEGER NOT NULL DEFAULT 0;

ALTER TABLE database_backup_runs ADD COLUMN sha256 TEXT NOT NULL DEFAULT '';
ALTER TABLE database_backup_runs ADD COLUMN verified_at INTEGER;

ALTER TABLE backup_targets ADD COLUMN managed_garage INTEGER NOT NULL DEFAULT 0;
ALTER TABLE backup_targets ADD COLUMN last_test_at INTEGER;
ALTER TABLE backup_targets ADD COLUMN last_test_error TEXT NOT NULL DEFAULT '';

CREATE TABLE database_restore_runs (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  backup_run_id TEXT NOT NULL REFERENCES database_backup_runs(id) ON DELETE CASCADE,
  mode TEXT NOT NULL CHECK (mode IN ('in_place', 'new_database')),
  status TEXT NOT NULL CHECK (status IN ('queued', 'running', 'succeeded', 'failed')),
  pre_snapshot_run_id TEXT REFERENCES database_backup_runs(id) ON DELETE SET NULL,
  target_service_id TEXT REFERENCES services(id) ON DELETE SET NULL,
  requested_by TEXT REFERENCES users(id) ON DELETE SET NULL,
  log_tail TEXT NOT NULL DEFAULT '',
  error TEXT NOT NULL DEFAULT '',
  started_at INTEGER,
  finished_at INTEGER,
  created_at INTEGER NOT NULL
);

CREATE INDEX idx_restore_runs_service ON database_restore_runs(service_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_restore_runs_service;
DROP TABLE IF EXISTS database_restore_runs;
ALTER TABLE backup_targets DROP COLUMN last_test_error;
ALTER TABLE backup_targets DROP COLUMN last_test_at;
ALTER TABLE backup_targets DROP COLUMN managed_garage;
ALTER TABLE database_backup_runs DROP COLUMN verified_at;
ALTER TABLE database_backup_runs DROP COLUMN sha256;
ALTER TABLE database_backup_policies DROP COLUMN run_weekday;
ALTER TABLE database_backup_policies DROP COLUMN run_at_minute;
