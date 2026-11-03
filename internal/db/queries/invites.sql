-- name: CreateInvite :exec
INSERT INTO invites (
  id,
  email,
  role,
  token_hash,
  invited_by,
  expires_at
) VALUES (?, ?, ?, ?, ?, ?);

-- name: ListPendingInvites :many
SELECT
  id,
  email,
  role,
  invited_by,
  expires_at
FROM invites
WHERE accepted_at IS NULL
  AND expires_at > ?
ORDER BY expires_at DESC;
