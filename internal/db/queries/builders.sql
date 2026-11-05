-- name: ListBuilders :many
SELECT
  id,
  name,
  kind,
  endpoint,
  platforms,
  status,
  last_seen_at,
  created_at
FROM builders
ORDER BY created_at DESC;

-- name: GetBuilder :one
SELECT
  id,
  name,
  kind,
  endpoint,
  ca_cert_enc,
  client_cert_enc,
  client_key_enc,
  platforms,
  status,
  last_seen_at,
  created_at
FROM builders
WHERE id = ?
LIMIT 1;

-- name: InsertBuilder :exec
INSERT INTO builders (
  id,
  name,
  kind,
  endpoint,
  ca_cert_enc,
  client_cert_enc,
  client_key_enc,
  platforms,
  status,
  last_seen_at,
  created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteBuilder :exec
DELETE FROM builders WHERE id = ?;

-- name: UpdateBuilderStatus :exec
UPDATE builders
SET status = ?, platforms = ?, last_seen_at = ?
WHERE id = ?;
