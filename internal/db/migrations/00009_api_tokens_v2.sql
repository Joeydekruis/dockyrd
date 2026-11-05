-- +goose Up
ALTER TABLE api_tokens ADD COLUMN project_id TEXT REFERENCES projects(id);
ALTER TABLE api_tokens ADD COLUMN revoked_at INTEGER;
ALTER TABLE api_tokens ADD COLUMN last_used_ip TEXT;

-- +goose Down
ALTER TABLE api_tokens DROP COLUMN last_used_ip;
ALTER TABLE api_tokens DROP COLUMN revoked_at;
ALTER TABLE api_tokens DROP COLUMN project_id;
