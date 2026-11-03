-- +goose Up
-- identiteit & toegang
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE COLLATE NOCASE,
  name TEXT NOT NULL,
  password_hash TEXT,
  role TEXT NOT NULL CHECK (role IN ('owner','admin','member','viewer')),
  totp_secret_enc BLOB,
  totp_enabled_at INTEGER,
  avatar_url TEXT,
  disabled_at INTEGER,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

CREATE TABLE sessions (
  id_hash BLOB PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  user_agent TEXT,
  ip TEXT,
  location TEXT,
  mfa_passed INTEGER NOT NULL DEFAULT 0,
  created_at INTEGER NOT NULL,
  last_seen_at INTEGER NOT NULL,
  expires_at INTEGER NOT NULL
);

CREATE TABLE recovery_codes (
  user_id TEXT,
  code_hash BLOB,
  used_at INTEGER
);

CREATE TABLE invites (
  id TEXT PRIMARY KEY,
  email TEXT,
  role TEXT,
  token_hash BLOB,
  invited_by TEXT,
  expires_at INTEGER,
  accepted_at INTEGER
);

CREATE TABLE password_resets (
  token_hash BLOB PRIMARY KEY,
  user_id TEXT,
  expires_at INTEGER,
  used_at INTEGER
);

CREATE TABLE oauth_identities (
  provider TEXT,
  provider_user_id TEXT,
  user_id TEXT,
  PRIMARY KEY (provider, provider_user_id)
);

CREATE TABLE project_members (
  project_id TEXT,
  user_id TEXT,
  role TEXT,
  PRIMARY KEY (project_id, user_id)
);

-- infrastructuur
CREATE TABLE servers (
  id TEXT PRIMARY KEY,
  name TEXT,
  kind TEXT DEFAULT 'local',
  public_ipv4 TEXT,
  public_ipv6 TEXT,
  created_at INTEGER
);

CREATE TABLE projects (
  id TEXT PRIMARY KEY,
  server_id TEXT NOT NULL,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  description TEXT,
  created_at INTEGER,
  updated_at INTEGER
);

CREATE TABLE environments (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  UNIQUE(project_id, name)
);

CREATE TABLE services (
  id TEXT PRIMARY KEY,
  environment_id TEXT NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  kind TEXT NOT NULL CHECK (kind IN ('app','static_site','postgres','redis','bucket')),
  name TEXT NOT NULL,
  slug TEXT NOT NULL,
  desired_state TEXT NOT NULL DEFAULT 'running',
  config JSON NOT NULL,
  live_deployment_id TEXT,
  created_at INTEGER,
  updated_at INTEGER,
  UNIQUE(environment_id, slug)
);

CREATE TABLE deployments (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  status TEXT NOT NULL,
  trigger TEXT NOT NULL,
  triggered_by TEXT,
  commit_sha TEXT,
  commit_message TEXT,
  commit_author TEXT,
  branch TEXT,
  image_ref TEXT,
  release_path TEXT,
  error TEXT,
  build_log_path TEXT,
  queued_at INTEGER,
  started_at INTEGER,
  finished_at INTEGER
);

CREATE TABLE domains (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  hostname TEXT NOT NULL UNIQUE COLLATE NOCASE,
  is_platform INTEGER NOT NULL DEFAULT 0,
  is_primary INTEGER NOT NULL DEFAULT 0,
  redirect_to TEXT,
  dns_status TEXT,
  dns_checked_at INTEGER,
  cert_status TEXT,
  cert_expires_at INTEGER
);

CREATE TABLE variables (
  id TEXT PRIMARY KEY,
  service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  key TEXT NOT NULL,
  value_enc BLOB NOT NULL,
  is_secret INTEGER NOT NULL DEFAULT 0,
  at_build INTEGER NOT NULL DEFAULT 1,
  at_runtime INTEGER NOT NULL DEFAULT 1,
  UNIQUE(service_id, key)
);

CREATE TABLE git_sources (
  id TEXT PRIMARY KEY,
  kind TEXT,
  github_app_id INTEGER,
  installation_id INTEGER,
  private_key_enc BLOB,
  webhook_secret_enc BLOB,
  client_id TEXT,
  client_secret_enc BLOB,
  slug TEXT,
  created_at INTEGER
);

-- werk & observability
CREATE TABLE jobs (
  id TEXT PRIMARY KEY,
  kind TEXT,
  payload JSON,
  status TEXT,
  attempts INTEGER DEFAULT 0,
  run_after INTEGER,
  locked_until INTEGER,
  last_error TEXT,
  created_at INTEGER
);

CREATE TABLE events (
  id TEXT PRIMARY KEY,
  level TEXT,
  category TEXT,
  title TEXT,
  detail TEXT,
  project_id TEXT,
  service_id TEXT,
  deployment_id TEXT,
  acknowledged_at INTEGER,
  created_at INTEGER
);

CREATE TABLE audit_log (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  action TEXT,
  target_kind TEXT,
  target_id TEXT,
  meta JSON,
  ip TEXT,
  created_at INTEGER
);

CREATE TABLE settings (
  key TEXT PRIMARY KEY,
  value JSON
);

CREATE TABLE metric_host (
  ts INTEGER,
  res INTEGER,
  cpu REAL,
  mem_used INTEGER,
  mem_total INTEGER,
  disk_used INTEGER,
  disk_total INTEGER,
  net_rx INTEGER,
  net_tx INTEGER,
  load1 REAL,
  PRIMARY KEY (res, ts)
) WITHOUT ROWID;

CREATE TABLE metric_container (
  ts INTEGER,
  res INTEGER,
  service_id TEXT,
  cpu REAL,
  mem INTEGER,
  mem_limit INTEGER,
  net_rx INTEGER,
  net_tx INTEGER,
  restarts INTEGER,
  PRIMARY KEY (service_id, res, ts)
) WITHOUT ROWID;

CREATE TABLE metric_http (
  ts INTEGER,
  res INTEGER,
  service_id TEXT,
  requests INTEGER,
  s2xx INTEGER,
  s3xx INTEGER,
  s4xx INTEGER,
  s5xx INTEGER,
  bytes INTEGER,
  p50 REAL,
  p90 REAL,
  p95 REAL,
  p99 REAL,
  PRIMARY KEY (service_id, res, ts)
) WITHOUT ROWID;

-- +goose Down
DROP TABLE IF EXISTS metric_http;
DROP TABLE IF EXISTS metric_container;
DROP TABLE IF EXISTS metric_host;
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS git_sources;
DROP TABLE IF EXISTS variables;
DROP TABLE IF EXISTS domains;
DROP TABLE IF EXISTS deployments;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS environments;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS servers;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS oauth_identities;
DROP TABLE IF EXISTS password_resets;
DROP TABLE IF EXISTS invites;
DROP TABLE IF EXISTS recovery_codes;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
