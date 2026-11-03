-- name: EnqueueJob :exec
INSERT INTO jobs (id, kind, payload, status, attempts, run_after, created_at)
VALUES (?, ?, ?, 'pending', 0, ?, ?);

-- name: ClaimNextJob :one
SELECT id, kind, payload, status, attempts, run_after, locked_until, last_error, created_at
FROM jobs
WHERE status = 'pending' AND run_after <= ?
ORDER BY created_at ASC
LIMIT 1;

-- name: LockJob :exec
UPDATE jobs
SET status = 'running', locked_until = ?, attempts = attempts + 1
WHERE id = ? AND status = 'pending';

-- name: CompleteJob :exec
UPDATE jobs
SET status = 'done', locked_until = NULL, last_error = NULL
WHERE id = ?;

-- name: FailJob :exec
UPDATE jobs
SET status = 'failed', locked_until = NULL, last_error = ?
WHERE id = ?;

-- name: RequeueJob :exec
UPDATE jobs
SET status = 'pending', locked_until = NULL, run_after = ?, last_error = ?
WHERE id = ?;
