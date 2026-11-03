-- name: GetSetting :one
SELECT key, value
FROM settings
WHERE key = ?;

-- name: UpsertSetting :exec
INSERT INTO settings (key, value)
VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value;

-- name: CountUsers :one
SELECT COUNT(*) AS count FROM users;
