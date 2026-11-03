-- name: InsertAuditLog :exec
INSERT INTO audit_log (
  id,
  user_id,
  action,
  target_kind,
  target_id,
  meta,
  ip,
  created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListAuditLogRecent :many
SELECT
  id,
  user_id,
  action,
  target_kind,
  target_id,
  meta,
  ip,
  created_at
FROM audit_log
ORDER BY created_at DESC
LIMIT ?;
