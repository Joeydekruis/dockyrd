-- name: DeleteAPITokensByScope :exec
DELETE FROM api_tokens WHERE scope = ?;

-- name: CountAPITokensByScope :one
SELECT COUNT(*) FROM api_tokens WHERE scope = ? AND revoked_at IS NULL;

-- name: InsertAPIToken :exec
INSERT INTO api_tokens (
  id, name, token_hash, scope, created_by, created_at, project_id
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetAPITokenByHash :one
SELECT id, name, token_hash, scope, created_by, last_used_at, expires_at, created_at,
       project_id, revoked_at, last_used_ip
FROM api_tokens
WHERE token_hash = ?
LIMIT 1;

-- name: ListAPITokensForUser :many
SELECT id, name, scope, project_id, last_used_at, last_used_ip, created_at, token_hash
FROM api_tokens
WHERE created_by = ? AND revoked_at IS NULL
ORDER BY created_at DESC;

-- name: RevokeAPIToken :exec
UPDATE api_tokens SET revoked_at = ? WHERE id = ? AND created_by = ? AND revoked_at IS NULL;

-- name: TouchAPITokenUsed :exec
UPDATE api_tokens SET last_used_at = ?, last_used_ip = ? WHERE id = ?;
