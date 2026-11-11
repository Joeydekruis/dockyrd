-- name: CreateSession :exec
INSERT INTO sessions (
  id_hash,
  user_id,
  user_agent,
  ip,
  mfa_passed,
  created_at,
  last_seen_at,
  expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetSession :one
SELECT user_id, expires_at, mfa_passed
FROM sessions
WHERE id_hash = ?;

-- name: SetSessionMFAPassed :exec
UPDATE sessions
SET mfa_passed = 1
WHERE id_hash = ?;

-- name: ListSessionsForUser :many
SELECT
  id_hash,
  user_agent,
  ip,
  mfa_passed,
  created_at,
  last_seen_at,
  expires_at
FROM sessions
WHERE user_id = ?
ORDER BY last_seen_at DESC;

-- name: DeleteSessionForUser :exec
DELETE FROM sessions
WHERE id_hash = ?
  AND user_id = ?;

-- name: DeleteAllSessionsForUser :exec
DELETE FROM sessions
WHERE user_id = ?;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id_hash = ?;

-- name: GetUserByID :one
SELECT id, email, name, role, totp_enabled_at
FROM users
WHERE id = ?
LIMIT 1;
