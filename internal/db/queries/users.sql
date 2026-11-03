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

-- name: ListUsers :many
SELECT
  id,
  email,
  name,
  role,
  created_at,
  disabled_at
FROM users
ORDER BY created_at ASC;

-- name: UpdateUserRole :exec
UPDATE users
SET role = ?, updated_at = ?
WHERE id = ?;
