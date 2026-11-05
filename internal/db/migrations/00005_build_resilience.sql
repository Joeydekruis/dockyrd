-- +goose Up
ALTER TABLE deployments ADD COLUMN mem_peak_bytes INTEGER;
ALTER TABLE deployments ADD COLUMN swap_peak_bytes INTEGER;
ALTER TABLE deployments ADD COLUMN cpu_seconds REAL;
ALTER TABLE deployments ADD COLUMN waited_ms INTEGER NOT NULL DEFAULT 0;
ALTER TABLE deployments ADD COLUMN frozen_ms INTEGER NOT NULL DEFAULT 0;
ALTER TABLE deployments ADD COLUMN budget_bytes INTEGER;
ALTER TABLE deployments ADD COLUMN builder TEXT NOT NULL DEFAULT 'local';
ALTER TABLE deployments ADD COLUMN failure_kind TEXT;

CREATE TABLE builders (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('remote_buildkit')),
  endpoint TEXT NOT NULL,
  ca_cert_enc BLOB,
  client_cert_enc BLOB,
  client_key_enc BLOB,
  platforms TEXT,
  status TEXT,
  last_seen_at INTEGER,
  created_at INTEGER NOT NULL
);

CREATE TABLE api_tokens (
  id TEXT PRIMARY KEY,
  name TEXT,
  token_hash BLOB NOT NULL,
  scope TEXT NOT NULL,
  created_by TEXT,
  last_used_at INTEGER,
  expires_at INTEGER,
  created_at INTEGER NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS builders;
ALTER TABLE deployments DROP COLUMN failure_kind;
ALTER TABLE deployments DROP COLUMN builder;
ALTER TABLE deployments DROP COLUMN budget_bytes;
ALTER TABLE deployments DROP COLUMN frozen_ms;
ALTER TABLE deployments DROP COLUMN waited_ms;
ALTER TABLE deployments DROP COLUMN cpu_seconds;
ALTER TABLE deployments DROP COLUMN swap_peak_bytes;
ALTER TABLE deployments DROP COLUMN mem_peak_bytes;
