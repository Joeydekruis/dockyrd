-- name: InsertEmailLog :exec
INSERT INTO email_log (
  id,
  kind,
  from_address,
  to_address,
  subject,
  status,
  response_text,
  detail_log,
  created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListEmailLogRecent :many
SELECT
  id,
  kind,
  from_address,
  to_address,
  subject,
  status,
  response_text,
  detail_log,
  created_at
FROM email_log
ORDER BY created_at DESC
LIMIT ?;

-- name: GetEmailLog :one
SELECT
  id,
  kind,
  from_address,
  to_address,
  subject,
  status,
  response_text,
  detail_log,
  created_at
FROM email_log
WHERE id = ?;
