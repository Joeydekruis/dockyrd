-- +goose Up
ALTER TABLE service_security ADD COLUMN waf_path_exceptions_json TEXT NOT NULL DEFAULT '[]';

CREATE TABLE security_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts TEXT NOT NULL,
  service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  source TEXT NOT NULL CHECK (source IN ('ratelimit', 'waf', 'crowdsec')),
  rule TEXT,
  path TEXT,
  action TEXT,
  ip TEXT,
  detail TEXT
);

CREATE INDEX idx_security_events_service_ts ON security_events (service_id, ts DESC);
CREATE INDEX idx_security_events_ts ON security_events (ts);

-- +goose Down
DROP INDEX IF EXISTS idx_security_events_ts;
DROP INDEX IF EXISTS idx_security_events_service_ts;
DROP TABLE IF EXISTS security_events;
-- SQLite cannot DROP COLUMN portably in older versions; leave waf_path_exceptions_json on down.
