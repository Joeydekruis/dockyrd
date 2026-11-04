-- name: InsertInboxNotification :exec
INSERT INTO inbox_notifications (
  id,
  user_id,
  category,
  level,
  title,
  body,
  source_label,
  cta_label,
  cta_href,
  email_token_hash,
  email_sent_at,
  read_at,
  deployment_id,
  service_id,
  project_id,
  created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListInboxNotificationsForUser :many
SELECT
  id,
  user_id,
  category,
  level,
  title,
  body,
  source_label,
  cta_label,
  cta_href,
  read_at,
  deployment_id,
  service_id,
  project_id,
  created_at
FROM inbox_notifications
WHERE user_id = ?
  AND (? = '' OR category = ?)
  AND (? = 0 OR read_at IS NULL)
  AND (? = 0 OR created_at < ?)
ORDER BY created_at DESC
LIMIT ?;

-- name: CountUnreadInboxForUser :one
SELECT COUNT(*) AS count
FROM inbox_notifications
WHERE user_id = ? AND read_at IS NULL;

-- name: CountInboxForUser :one
SELECT COUNT(*) AS count
FROM inbox_notifications
WHERE user_id = ?;

-- name: GetInboxNotificationForUser :one
SELECT
  id,
  user_id,
  category,
  level,
  title,
  body,
  source_label,
  cta_label,
  cta_href,
  read_at,
  deployment_id,
  service_id,
  project_id,
  created_at
FROM inbox_notifications
WHERE id = ? AND user_id = ?
LIMIT 1;

-- name: MarkInboxNotificationRead :exec
UPDATE inbox_notifications
SET read_at = ?
WHERE id = ? AND user_id = ? AND read_at IS NULL;

-- name: MarkAllInboxNotificationsRead :exec
UPDATE inbox_notifications
SET read_at = ?
WHERE user_id = ? AND read_at IS NULL;

-- name: GetInboxNotificationByEmailToken :one
SELECT
  id,
  user_id,
  category,
  level,
  title,
  body,
  source_label,
  cta_label,
  cta_href,
  read_at,
  deployment_id,
  service_id,
  project_id,
  created_at
FROM inbox_notifications
WHERE email_token_hash = ?
LIMIT 1;

-- name: MarkInboxNotificationReadByID :exec
UPDATE inbox_notifications
SET read_at = ?
WHERE id = ? AND read_at IS NULL;

-- name: SetInboxNotificationEmailSent :exec
UPDATE inbox_notifications
SET email_sent_at = ?
WHERE id = ?;
