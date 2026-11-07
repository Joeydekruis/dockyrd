-- name: InsertHostAction :exec
INSERT INTO host_actions (
  id,
  kind,
  state,
  plan_json,
  revert_json,
  actor_user_id,
  applied_at,
  confirm_deadline,
  confirmed_at,
  reverted_at,
  last_error,
  created_at,
  updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetHostAction :one
SELECT
  id,
  kind,
  state,
  plan_json,
  revert_json,
  actor_user_id,
  applied_at,
  confirm_deadline,
  confirmed_at,
  reverted_at,
  last_error,
  created_at,
  updated_at
FROM host_actions
WHERE id = ?;

-- name: UpdateHostAction :execrows
UPDATE host_actions
SET
  state = ?,
  plan_json = ?,
  revert_json = ?,
  actor_user_id = ?,
  applied_at = ?,
  confirm_deadline = ?,
  confirmed_at = ?,
  reverted_at = ?,
  last_error = ?,
  updated_at = ?
WHERE id = sqlc.arg(id) AND state = sqlc.arg(expected_state);

-- name: ListHostActionsAwaitingConfirmExpired :many
SELECT
  id,
  kind,
  state,
  plan_json,
  revert_json,
  actor_user_id,
  applied_at,
  confirm_deadline,
  confirmed_at,
  reverted_at,
  last_error,
  created_at,
  updated_at
FROM host_actions
WHERE state = 'awaiting_confirm' AND confirm_deadline < ?;

-- name: ListHostActionsInFlight :many
SELECT
  id,
  kind,
  state,
  plan_json,
  revert_json,
  actor_user_id,
  applied_at,
  confirm_deadline,
  confirmed_at,
  reverted_at,
  last_error,
  created_at,
  updated_at
FROM host_actions
WHERE state IN ('planned', 'applied', 'verified', 'awaiting_confirm', 'failed')
ORDER BY updated_at DESC;
