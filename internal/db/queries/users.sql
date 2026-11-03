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
  disabled_at,
  totp_enabled_at
FROM users
WHERE email = ? COLLATE NOCASE
LIMIT 1;

-- name: GetUserSecurity :one
SELECT
  id,
  email,
  name,
  role,
  totp_secret_enc,
  totp_enabled_at
FROM users
WHERE id = ?
LIMIT 1;

-- name: EnableUserTOTP :exec
UPDATE users
SET
  totp_secret_enc = ?,
  totp_enabled_at = ?,
  updated_at = ?
WHERE id = ?;

-- name: DisableUserTOTP :exec
UPDATE users
SET
  totp_secret_enc = NULL,
  totp_enabled_at = NULL,
  updated_at = ?
WHERE id = ?;

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

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = ?, updated_at = ?
WHERE id = ?;
