-- +goose Up
CREATE TABLE inbox_notifications (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  category TEXT NOT NULL CHECK (category IN ('deploys', 'usage', 'system')),
  level TEXT NOT NULL CHECK (level IN ('error', 'warning', 'success', 'info')),
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  source_label TEXT NOT NULL DEFAULT '',
  cta_label TEXT,
  cta_href TEXT,
  email_token_hash BLOB,
  email_sent_at INTEGER,
  read_at INTEGER,
  deployment_id TEXT,
  service_id TEXT,
  project_id TEXT,
  created_at INTEGER NOT NULL
);

CREATE INDEX idx_inbox_notifications_user_created
  ON inbox_notifications (user_id, created_at DESC);

CREATE INDEX idx_inbox_notifications_user_unread
  ON inbox_notifications (user_id, read_at, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_inbox_notifications_user_unread;
DROP INDEX IF EXISTS idx_inbox_notifications_user_created;
DROP TABLE IF EXISTS inbox_notifications;
