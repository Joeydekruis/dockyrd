-- name: InsertPasswordReset :exec
INSERT INTO password_resets (token_hash, user_id, expires_at)
VALUES (?, ?, ?);

-- name: GetPasswordReset :one
SELECT user_id, expires_at, used_at
FROM password_resets
WHERE token_hash = ?
LIMIT 1;

-- name: MarkPasswordResetUsed :exec
UPDATE password_resets
SET used_at = ?
WHERE token_hash = ?;
