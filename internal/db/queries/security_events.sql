-- name: InsertSecurityEvent :one
INSERT INTO security_events (
  ts, service_id, source, rule, path, action, ip, detail
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING id, ts, service_id, source, rule, path, action, ip, detail;

-- name: ListSecurityEventsByService :many
SELECT id, ts, service_id, source, rule, path, action, ip, detail
FROM security_events
WHERE service_id = ?
ORDER BY ts DESC, id DESC
LIMIT ?;

-- name: ListSecurityEventsByServiceSource :many
SELECT id, ts, service_id, source, rule, path, action, ip, detail
FROM security_events
WHERE service_id = ? AND source = ?
ORDER BY ts DESC, id DESC
LIMIT ?;

-- name: PurgeSecurityEventsBefore :execrows
DELETE FROM security_events
WHERE ts < ?;
