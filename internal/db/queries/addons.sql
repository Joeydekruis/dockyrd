-- name: ListAddons :many
SELECT id, state, version, config_json, installed_at, last_error
FROM addons
ORDER BY id;

-- name: GetAddon :one
SELECT id, state, version, config_json, installed_at, last_error
FROM addons
WHERE id = ?;

-- name: UpsertAddon :exec
INSERT INTO addons (id, state, version, config_json, installed_at, last_error)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  state = excluded.state,
  version = excluded.version,
  config_json = excluded.config_json,
  installed_at = excluded.installed_at,
  last_error = excluded.last_error;

-- name: UpdateAddonState :execrows
UPDATE addons
SET state = ?, last_error = ?, version = ?, installed_at = ?, config_json = ?
WHERE id = ? AND state = ?;
