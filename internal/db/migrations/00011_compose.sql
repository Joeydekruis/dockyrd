-- +goose Up
-- +goose NO TRANSACTION
-- Rebuild services to add kind=compose; extend domains for compose upstreams.
PRAGMA foreign_keys = OFF;
BEGIN;

CREATE TABLE services_new (
  id TEXT PRIMARY KEY,
  environment_id TEXT NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  kind TEXT NOT NULL CHECK (kind IN ('app','static_site','postgres','redis','bucket','compose')),
  name TEXT NOT NULL,
  slug TEXT NOT NULL,
  desired_state TEXT NOT NULL DEFAULT 'running',
  config JSON NOT NULL,
  live_deployment_id TEXT,
  created_at INTEGER,
  updated_at INTEGER,
  parent_service_id TEXT REFERENCES services(id) ON DELETE CASCADE,
  preview_pr INTEGER,
  preview_head_sha TEXT,
  preview_last_push_at INTEGER,
  UNIQUE(environment_id, slug)
);

INSERT INTO services_new (
  id, environment_id, kind, name, slug, desired_state, config,
  live_deployment_id, created_at, updated_at,
  parent_service_id, preview_pr, preview_head_sha, preview_last_push_at
)
SELECT
  id, environment_id, kind, name, slug, desired_state, config,
  live_deployment_id, created_at, updated_at,
  parent_service_id, preview_pr, preview_head_sha, preview_last_push_at
FROM services;

DROP TABLE services;
ALTER TABLE services_new RENAME TO services;

CREATE UNIQUE INDEX idx_services_preview_pr ON services(parent_service_id, preview_pr)
WHERE parent_service_id IS NOT NULL;

COMMIT;
PRAGMA foreign_key_check;
PRAGMA foreign_keys = ON;

ALTER TABLE domains ADD COLUMN target_service TEXT;
ALTER TABLE domains ADD COLUMN target_port INTEGER;

-- +goose Down
ALTER TABLE domains DROP COLUMN target_port;
ALTER TABLE domains DROP COLUMN target_service;

PRAGMA foreign_keys = OFF;
BEGIN;

CREATE TABLE services_old (
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
  parent_service_id TEXT REFERENCES services(id) ON DELETE CASCADE,
  preview_pr INTEGER,
  preview_head_sha TEXT,
  preview_last_push_at INTEGER,
  UNIQUE(environment_id, slug)
);

INSERT INTO services_old SELECT * FROM services WHERE kind != 'compose';

DROP TABLE services;
ALTER TABLE services_old RENAME TO services;

CREATE UNIQUE INDEX idx_services_preview_pr ON services(parent_service_id, preview_pr)
WHERE parent_service_id IS NOT NULL;

COMMIT;
PRAGMA foreign_keys = ON;
