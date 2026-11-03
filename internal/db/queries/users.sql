-- name: CreateUser :exec
INSERT INTO users (
  id,
  email,
  name,
  password_hash,
  role,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetUserByEmail :one
SELECT
  id,
  email,
  name,
  password_hash,
  role,
  disabled_at
FROM users
WHERE email = ? COLLATE NOCASE
LIMIT 1;
