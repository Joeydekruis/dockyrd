-- name: CreateServiceLink :exec
INSERT INTO service_links (id, from_service_id, to_service_id, created_by, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetServiceLinkByID :one
SELECT id, from_service_id, to_service_id, created_by, created_at
FROM service_links
WHERE id = ?;

-- name: GetServiceLinkByPair :one
SELECT id, from_service_id, to_service_id, created_by, created_at
FROM service_links
WHERE from_service_id = ? AND to_service_id = ?;

-- name: DeleteServiceLink :exec
DELETE FROM service_links
WHERE id = ?;

-- name: ListOutgoingLinks :many
SELECT id, from_service_id, to_service_id, created_by, created_at
FROM service_links
WHERE from_service_id = ?
ORDER BY created_at ASC;

-- name: ListIncomingLinks :many
SELECT id, from_service_id, to_service_id, created_by, created_at
FROM service_links
WHERE to_service_id = ?
ORDER BY created_at ASC;

-- name: ListLinksForProject :many
SELECT sl.id, sl.from_service_id, sl.to_service_id, sl.created_by, sl.created_at
FROM service_links sl
JOIN services fs ON fs.id = sl.from_service_id
JOIN environments fe ON fe.id = fs.environment_id
WHERE fe.project_id = ?
ORDER BY sl.created_at ASC;

-- name: ListLinkedProviderSlugs :many
SELECT s.slug
FROM service_links sl
JOIN services s ON s.id = sl.to_service_id
WHERE sl.from_service_id = ?;

-- name: ListVariablesByLink :many
SELECT id, service_id, key, value_enc, is_secret, at_build, at_runtime, link_id
FROM variables
WHERE link_id = ?;

-- name: DeleteVariablesByLink :exec
DELETE FROM variables
WHERE link_id = ?;

-- name: InsertVariableWithLink :exec
INSERT INTO variables (id, service_id, key, value_enc, is_secret, at_build, at_runtime, link_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateProjectLinksMode :exec
UPDATE projects
SET links_mode = ?, updated_at = ?
WHERE id = ?;

-- name: GetProjectLinksMode :one
SELECT links_mode
FROM projects
WHERE id = ?;
