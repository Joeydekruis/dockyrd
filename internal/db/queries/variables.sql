-- name: ListVariablesForService :many
SELECT id, service_id, key, value_enc, is_secret, at_build, at_runtime, link_id
FROM variables
WHERE service_id = ?
ORDER BY key ASC;

-- name: DeleteVariablesForService :exec
DELETE FROM variables
WHERE service_id = ?;

-- name: InsertVariable :exec
INSERT INTO variables (id, service_id, key, value_enc, is_secret, at_build, at_runtime)
VALUES (?, ?, ?, ?, ?, ?, ?);
