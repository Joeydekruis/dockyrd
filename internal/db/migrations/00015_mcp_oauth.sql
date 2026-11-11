-- +goose Up
CREATE TABLE mcp_oauth_codes (
  code_hash BLOB PRIMARY KEY NOT NULL,
  client_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  redirect_uri TEXT NOT NULL,
  code_challenge TEXT NOT NULL,
  scope TEXT NOT NULL,
  expires_at INTEGER NOT NULL
);

CREATE TABLE mcp_access_tokens (
  token_hash BLOB PRIMARY KEY NOT NULL,
  user_id TEXT NOT NULL,
  scope TEXT NOT NULL,
  expires_at INTEGER NOT NULL,
  created_at INTEGER NOT NULL
);

CREATE INDEX idx_mcp_access_tokens_expires ON mcp_access_tokens(expires_at);

-- +goose Down
DROP TABLE IF EXISTS mcp_access_tokens;
DROP TABLE IF EXISTS mcp_oauth_codes;
