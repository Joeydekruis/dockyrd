-- name: InsertTypedEvent :exec
INSERT INTO events (
  id,
  level,
  category,
  title,
  detail,
  project_id,
  service_id,
  deployment_id,
  acknowledged_at,
  created_at,
  type,
  state,
  dedup_key,
  data
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListNotificationChannels :many
SELECT
  id,
  kind,
  name,
  enabled,
  config,
  secret_enc,
  min_level,
  event_types,
  project_id,
  status,
  disabled_reason,
  consecutive_failures,
  last_success_at,
  last_failure_at,
  last_error,
  created_by,
  created_at,
  updated_at
FROM notification_channels
ORDER BY created_at DESC;

-- name: GetNotificationChannel :one
SELECT
  id,
  kind,
  name,
  enabled,
  config,
  secret_enc,
  min_level,
  event_types,
  project_id,
  status,
  disabled_reason,
  consecutive_failures,
  last_success_at,
  last_failure_at,
  last_error,
  created_by,
  created_at,
  updated_at
FROM notification_channels
WHERE id = ?
LIMIT 1;

-- name: InsertNotificationChannel :exec
INSERT INTO notification_channels (
  id,
  kind,
  name,
  enabled,
  config,
  secret_enc,
  min_level,
  event_types,
  project_id,
  status,
  disabled_reason,
  consecutive_failures,
  last_success_at,
  last_failure_at,
  last_error,
  created_by,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateNotificationChannel :exec
UPDATE notification_channels
SET
  name = ?,
  enabled = ?,
  config = ?,
  secret_enc = ?,
  min_level = ?,
  event_types = ?,
  project_id = ?,
  status = ?,
  disabled_reason = ?,
  consecutive_failures = ?,
  last_success_at = ?,
  last_failure_at = ?,
  last_error = ?,
  updated_at = ?
WHERE id = ?;

-- name: DeleteNotificationChannel :exec
DELETE FROM notification_channels WHERE id = ?;

-- name: CountNotificationChannels :one
SELECT COUNT(*) AS count FROM notification_channels;

-- name: GetAlertState :one
SELECT
  dedup_key,
  type,
  state,
  level,
  first_fired_at,
  last_notified_at,
  resolved_at,
  occurrences,
  event_id
FROM alert_state
WHERE dedup_key = ?
LIMIT 1;

-- name: UpsertAlertState :exec
INSERT INTO alert_state (
  dedup_key,
  type,
  state,
  level,
  first_fired_at,
  last_notified_at,
  resolved_at,
  occurrences,
  event_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(dedup_key) DO UPDATE SET
  type = excluded.type,
  state = excluded.state,
  level = excluded.level,
  last_notified_at = excluded.last_notified_at,
  resolved_at = excluded.resolved_at,
  occurrences = excluded.occurrences,
  event_id = excluded.event_id;

-- name: InsertNotificationDelivery :exec
INSERT INTO notification_deliveries (
  id,
  event_id,
  channel_id,
  target,
  dedup_key,
  status,
  attempts,
  next_attempt_at,
  lease_until,
  http_status,
  last_error,
  created_at,
  sent_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: SkipPendingDeliveriesByDedup :exec
UPDATE notification_deliveries
SET status = 'skipped', last_error = ?
WHERE channel_id = ?
  AND dedup_key = ?
  AND status = 'pending';

-- name: RecoverExpiredDeliveryLeases :exec
UPDATE notification_deliveries
SET status = 'pending', lease_until = NULL
WHERE status = 'sending'
  AND lease_until IS NOT NULL
  AND lease_until < ?;

-- name: ListDueDeliveries :many
SELECT
  d.id,
  d.event_id,
  d.channel_id,
  d.target,
  d.attempts
FROM notification_deliveries d
JOIN notification_channels c ON c.id = d.channel_id
WHERE d.status = 'pending'
  AND d.next_attempt_at <= ?
  AND c.status != 'disabled'
  AND c.enabled = 1
ORDER BY d.next_attempt_at ASC, d.created_at ASC
LIMIT ?;

-- name: MarkDeliverySending :exec
UPDATE notification_deliveries
SET
  status = 'sending',
  lease_until = ?,
  attempts = attempts + 1
WHERE id = ?
  AND status = 'pending';

-- name: MarkDeliverySent :exec
UPDATE notification_deliveries
SET
  status = 'sent',
  sent_at = ?,
  http_status = ?,
  last_error = NULL,
  lease_until = NULL
WHERE id = ?;

-- name: MarkDeliveryRetry :exec
UPDATE notification_deliveries
SET
  status = 'pending',
  next_attempt_at = ?,
  http_status = ?,
  last_error = ?,
  lease_until = NULL
WHERE id = ?;

-- name: MarkDeliveryDead :exec
UPDATE notification_deliveries
SET
  status = 'dead',
  http_status = ?,
  last_error = ?,
  lease_until = NULL
WHERE id = ?;

-- name: UpdateChannelDeliveryHealth :exec
UPDATE notification_channels
SET
  consecutive_failures = ?,
  last_success_at = ?,
  last_failure_at = ?,
  last_error = ?,
  status = ?,
  disabled_reason = ?,
  updated_at = ?
WHERE id = ?;

-- name: ListChannelDeliveries :many
SELECT
  id,
  event_id,
  channel_id,
  target,
  dedup_key,
  status,
  attempts,
  next_attempt_at,
  lease_until,
  http_status,
  last_error,
  created_at,
  sent_at
FROM notification_deliveries
WHERE channel_id = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: GetNotificationDelivery :one
SELECT
  id,
  event_id,
  channel_id,
  target,
  dedup_key,
  status,
  attempts,
  next_attempt_at,
  lease_until,
  http_status,
  last_error,
  created_at,
  sent_at
FROM notification_deliveries
WHERE id = ?
LIMIT 1;

-- name: RetryNotificationDelivery :exec
UPDATE notification_deliveries
SET
  status = 'pending',
  next_attempt_at = ?,
  last_error = NULL,
  lease_until = NULL
WHERE id = ?
  AND status = 'dead';

-- name: GetEventForNotify :one
SELECT
  id,
  level,
  category,
  title,
  detail,
  project_id,
  service_id,
  deployment_id,
  acknowledged_at,
  created_at,
  type,
  state,
  dedup_key,
  data
FROM events
WHERE id = ?
LIMIT 1;
