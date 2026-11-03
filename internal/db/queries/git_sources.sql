-- name: GetGitSource :one
SELECT
  id,
  kind,
  github_app_id,
  installation_id,
  private_key_enc,
  webhook_secret_enc,
  client_id,
  client_secret_enc,
  slug,
  created_at
FROM git_sources
WHERE id = ?
LIMIT 1;

-- name: UpsertGitSource :exec
INSERT INTO git_sources (
  id,
  kind,
  github_app_id,
  installation_id,
  private_key_enc,
  webhook_secret_enc,
  client_id,
  client_secret_enc,
  slug,
  created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  kind = excluded.kind,
  github_app_id = excluded.github_app_id,
  installation_id = excluded.installation_id,
  private_key_enc = excluded.private_key_enc,
  webhook_secret_enc = excluded.webhook_secret_enc,
  client_id = excluded.client_id,
  client_secret_enc = excluded.client_secret_enc,
  slug = excluded.slug;

-- name: SetGitSourceInstallation :exec
UPDATE git_sources
SET installation_id = ?
WHERE id = ?;
