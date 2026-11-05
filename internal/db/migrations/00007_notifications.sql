-- +goose Up
CREATE TABLE notification_channels (
  id                   TEXT PRIMARY KEY,
  kind                 TEXT NOT NULL CHECK (kind IN ('slack','discord','email','webhook')),
  name                 TEXT NOT NULL,
  enabled              INTEGER NOT NULL DEFAULT 1,
  config               TEXT NOT NULL DEFAULT '{}',
  secret_enc           BLOB,
  min_level            TEXT NOT NULL DEFAULT 'warning' CHECK (min_level IN ('info','warning','error')),
  event_types          TEXT NOT NULL DEFAULT '[]',
  project_id           TEXT REFERENCES projects(id) ON DELETE CASCADE,
  status               TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','degraded','disabled')),
  disabled_reason      TEXT,
  consecutive_failures INTEGER NOT NULL DEFAULT 0,
  last_success_at      INTEGER,
  last_failure_at      INTEGER,
  last_error           TEXT,
  created_by           TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_at           INTEGER NOT NULL,
  updated_at           INTEGER NOT NULL
);

ALTER TABLE events ADD COLUMN type      TEXT;
ALTER TABLE events ADD COLUMN state     TEXT;
ALTER TABLE events ADD COLUMN dedup_key TEXT;
ALTER TABLE events ADD COLUMN data      TEXT;
CREATE INDEX idx_events_type_created ON events (type, created_at DESC);

CREATE TABLE alert_state (
  dedup_key        TEXT PRIMARY KEY,
  type             TEXT NOT NULL,
  state            TEXT NOT NULL CHECK (state IN ('firing','resolved')),
  level            TEXT NOT NULL,
  first_fired_at   INTEGER NOT NULL,
  last_notified_at INTEGER NOT NULL,
  resolved_at      INTEGER,
  occurrences      INTEGER NOT NULL DEFAULT 1,
  event_id         TEXT
);

CREATE TABLE notification_deliveries (
  id              TEXT PRIMARY KEY,
  event_id        TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  channel_id      TEXT NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
  target          TEXT NOT NULL DEFAULT '',
  dedup_key       TEXT,
  status          TEXT NOT NULL CHECK (status IN ('pending','sending','sent','dead','skipped')),
  attempts        INTEGER NOT NULL DEFAULT 0,
  next_attempt_at INTEGER NOT NULL,
  lease_until     INTEGER,
  http_status     INTEGER,
  last_error      TEXT,
  created_at      INTEGER NOT NULL,
  sent_at         INTEGER,
  UNIQUE (event_id, channel_id, target)
);
CREATE INDEX idx_deliveries_due     ON notification_deliveries (status, next_attempt_at);
CREATE INDEX idx_deliveries_channel ON notification_deliveries (channel_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_deliveries_channel;
DROP INDEX IF EXISTS idx_deliveries_due;
DROP TABLE IF EXISTS notification_deliveries;
DROP TABLE IF EXISTS alert_state;
DROP INDEX IF EXISTS idx_events_type_created;
ALTER TABLE events DROP COLUMN data;
ALTER TABLE events DROP COLUMN dedup_key;
ALTER TABLE events DROP COLUMN state;
ALTER TABLE events DROP COLUMN type;
DROP TABLE IF EXISTS notification_channels;
