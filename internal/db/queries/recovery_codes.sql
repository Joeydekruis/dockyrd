-- name: DeleteRecoveryCodesForUser :exec
DELETE FROM recovery_codes
WHERE user_id = ?;

-- name: InsertRecoveryCode :exec
INSERT INTO recovery_codes (user_id, code_hash)
VALUES (?, ?);

-- name: ConsumeRecoveryCode :exec
UPDATE recovery_codes
SET used_at = ?
WHERE user_id = ?
  AND code_hash = ?
  AND used_at IS NULL;

-- name: RecoveryCodeValid :one
SELECT user_id
FROM recovery_codes
WHERE user_id = ?
  AND code_hash = ?
  AND used_at IS NULL
LIMIT 1;
