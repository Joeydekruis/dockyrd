-- name: UpdateDomainCertStatus :exec
UPDATE domains
SET cert_status = ?, cert_expires_at = ?
WHERE id = ?;

-- name: ListDomainsForCertWatch :many
SELECT
  d.id,
  d.hostname,
  d.dns_status,
  d.service_id,
  s.desired_state
FROM domains d
JOIN services s ON s.id = d.service_id
WHERE d.is_platform = 0;

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

-- name: ClearPrimaryForService :exec
UPDATE domains SET is_primary = 0 WHERE service_id = ?;

-- name: SetDomainPrimary :exec
UPDATE domains SET is_primary = 1 WHERE id = ?;

-- name: SetDomainPrimaryForService :exec
UPDATE domains
SET is_primary = CASE WHEN id = ? THEN 1 ELSE 0 END
WHERE service_id = ?;

-- name: ListAppDomainsWithUpstream :many
SELECT
  d.hostname,
  s.slug AS service_slug,
  s.config AS service_config
FROM domains d
JOIN services s ON s.id = d.service_id
WHERE s.kind = 'app' AND s.desired_state = 'running';
