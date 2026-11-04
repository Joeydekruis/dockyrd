-- +goose Up
CREATE TABLE email_log (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  from_address TEXT,
  to_address TEXT NOT NULL,
  subject TEXT,
  status TEXT NOT NULL CHECK (status IN ('sent', 'failed')),
  response_text TEXT,
  detail_log TEXT,
  created_at INTEGER NOT NULL
);

CREATE INDEX idx_email_log_created_at ON email_log (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS email_log;
