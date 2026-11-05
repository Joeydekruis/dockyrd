-- +goose Up
CREATE TABLE service_links (
  id TEXT PRIMARY KEY,
  from_service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  to_service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  created_by TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_at INTEGER NOT NULL,
  UNIQUE (from_service_id, to_service_id),
  CHECK (from_service_id <> to_service_id)
);

CREATE INDEX idx_service_links_to ON service_links(to_service_id);

ALTER TABLE variables ADD COLUMN link_id TEXT REFERENCES service_links(id) ON DELETE SET NULL;

ALTER TABLE projects ADD COLUMN links_mode TEXT NOT NULL DEFAULT 'open' CHECK (links_mode IN ('open', 'strict'));

-- +goose Down
ALTER TABLE projects DROP COLUMN links_mode;
-- SQLite cannot drop column with FK easily in older versions; forward-only in prod.
-- For dev down: recreate variables without link_id if needed.
