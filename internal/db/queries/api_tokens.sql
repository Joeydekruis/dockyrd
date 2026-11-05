-- name: DeleteAPITokensByScope :exec
DELETE FROM api_tokens WHERE scope = ?;

-- name: InsertAPIToken :exec
INSERT INTO api_tokens (
  id, name, token_hash, scope, created_by, created_at
) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetAPITokenByHash :one
SELECT id, name, token_hash, scope, created_by, last_used_at, expires_at, created_at
FROM api_tokens
WHERE token_hash = ?
LIMIT 1;

-- name: TouchAPITokenUsed :exec
UPDATE api_tokens SET last_used_at = ? WHERE id = ?;
