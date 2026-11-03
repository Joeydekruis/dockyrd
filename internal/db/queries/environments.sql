-- name: CreateEnvironment :exec
INSERT INTO environments (id, project_id, name)
VALUES (?, ?, ?);

-- name: ListEnvironmentsForProject :many
SELECT id, project_id, name
FROM environments
WHERE project_id = ?
ORDER BY name ASC;
