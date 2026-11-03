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
) VALUES (?, ?, ?, ?, 1, ?, ?, ?);

-- name: GetSession :one
SELECT user_id, expires_at
FROM sessions
WHERE id_hash = ?;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id_hash = ?;

-- name: GetUserByID :one
SELECT id, email, name, role
FROM users
WHERE id = ?
LIMIT 1;
