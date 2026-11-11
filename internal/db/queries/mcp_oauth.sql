-- name: InsertMCPOAuthCode :exec
INSERT INTO mcp_oauth_codes (
  code_hash, client_id, user_id, redirect_uri, code_challenge, scope, expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetMCPOAuthCode :one
SELECT code_hash, client_id, user_id, redirect_uri, code_challenge, scope, expires_at
FROM mcp_oauth_codes
WHERE code_hash = ?
LIMIT 1;

-- name: DeleteMCPOAuthCode :exec
DELETE FROM mcp_oauth_codes WHERE code_hash = ?;

-- name: InsertMCPAccessToken :exec
INSERT INTO mcp_access_tokens (token_hash, user_id, scope, expires_at, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetMCPAccessTokenByHash :one
SELECT token_hash, user_id, scope, expires_at, created_at
FROM mcp_access_tokens
WHERE token_hash = ?
LIMIT 1;

-- name: DeleteExpiredMCPAccessTokens :exec
DELETE FROM mcp_access_tokens WHERE expires_at < ?;
