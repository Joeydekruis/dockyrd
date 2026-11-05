-- +goose Up
ALTER TABLE services ADD COLUMN parent_service_id TEXT REFERENCES services(id) ON DELETE CASCADE;
ALTER TABLE services ADD COLUMN preview_pr INTEGER;
ALTER TABLE services ADD COLUMN preview_head_sha TEXT;
ALTER TABLE services ADD COLUMN preview_last_push_at INTEGER;

CREATE UNIQUE INDEX idx_services_preview_pr ON services(parent_service_id, preview_pr)
WHERE parent_service_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_services_preview_pr;
ALTER TABLE services DROP COLUMN preview_last_push_at;
ALTER TABLE services DROP COLUMN preview_head_sha;
ALTER TABLE services DROP COLUMN preview_pr;
ALTER TABLE services DROP COLUMN parent_service_id;
