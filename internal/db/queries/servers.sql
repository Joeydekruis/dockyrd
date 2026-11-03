-- name: GetServerByID :one
SELECT id, name, kind, public_ipv4, public_ipv6, created_at
FROM servers
WHERE id = ?
LIMIT 1;

-- name: GetLocalServer :one
SELECT id, name, kind, public_ipv4, public_ipv6, created_at
FROM servers
WHERE kind = 'local'
ORDER BY created_at ASC
LIMIT 1;

-- name: CreateServer :exec
INSERT INTO servers (id, name, kind, public_ipv4, public_ipv6, created_at)
VALUES (?, ?, ?, ?, ?, ?);
