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
ORDER BY expires_at DESC;

-- name: GetPendingInviteByTokenHash :one
SELECT
  id,
  email,
  role,
  expires_at
FROM invites
WHERE token_hash = ?
  AND accepted_at IS NULL
  AND expires_at > ?
LIMIT 1;

-- name: MarkInviteAccepted :exec
UPDATE invites
SET accepted_at = ?
WHERE id = ?;

-- name: DeleteInvite :exec
DELETE FROM invites
WHERE id = ? AND accepted_at IS NULL;
