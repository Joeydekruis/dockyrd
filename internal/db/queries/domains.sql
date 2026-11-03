-- name: ListDomainsForService :many
SELECT
  id,
  service_id,
  hostname,
  is_platform,
  is_primary,
  redirect_to,
  dns_status,
  dns_checked_at,
  cert_status,
  cert_expires_at
FROM domains
WHERE service_id = ?
ORDER BY is_primary DESC, hostname ASC;

-- name: GetDomainByID :one
SELECT
  id,
  service_id,
  hostname,
  is_platform,
  is_primary,
  redirect_to,
  dns_status,
  dns_checked_at,
  cert_status,
  cert_expires_at
FROM domains
WHERE id = ?
LIMIT 1;

-- name: CreateDomain :exec
INSERT INTO domains (
  id,
  service_id,
  hostname,
  is_platform,
  is_primary,
  redirect_to,
  dns_status,
  dns_checked_at,
  cert_status,
  cert_expires_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteDomain :exec
DELETE FROM domains
WHERE id = ?;

-- name: ListAppDomainsWithUpstream :many
SELECT
  d.hostname,
  s.slug AS service_slug,
  s.config AS service_config
FROM domains d
JOIN services s ON s.id = d.service_id
WHERE s.kind = 'app' AND s.desired_state = 'running';
