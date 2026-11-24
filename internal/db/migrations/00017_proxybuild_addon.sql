-- +goose Up
INSERT OR IGNORE INTO addons (id, state) VALUES ('proxybuild', 'not_installed');

-- +goose Down
DELETE FROM addons WHERE id = 'proxybuild';