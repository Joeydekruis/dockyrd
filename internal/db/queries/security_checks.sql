-- name: GetSecurityCheckCache :one
SELECT ran_at, results_json
FROM security_check_cache
WHERE id = 1;

-- name: UpsertSecurityCheckCache :exec
INSERT INTO security_check_cache (id, ran_at, results_json)
VALUES (1, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  ran_at = excluded.ran_at,
  results_json = excluded.results_json;
