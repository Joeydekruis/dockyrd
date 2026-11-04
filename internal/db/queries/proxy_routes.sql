-- name: GetServiceDeployContext :one
SELECT
  s.id AS service_id,
  s.kind AS service_kind,
  s.slug AS service_slug,
  s.config AS service_config,
  s.live_deployment_id,
  e.project_id,
  p.slug AS project_slug,
  p.server_id,
  srv.public_ipv4
FROM services s
JOIN environments e ON e.id = s.environment_id
JOIN projects p ON p.id = e.project_id
JOIN servers srv ON srv.id = p.server_id
WHERE s.id = ?
LIMIT 1;

-- name: ListLiveAppProxyRoutes :many
SELECT
  d.hostname,
  s.slug AS service_slug,
  dep.id AS deployment_id,
  s.config AS service_config
FROM domains d
JOIN services s ON s.id = d.service_id
JOIN deployments dep ON dep.id = s.live_deployment_id
WHERE s.kind = 'app' AND dep.status = 'live';

-- name: ListLiveStaticProxyRoutes :many
SELECT
  d.hostname,
  s.slug AS service_slug,
  s.config AS service_config,
  dep.release_path
FROM domains d
JOIN services s ON s.id = d.service_id
JOIN deployments dep ON dep.id = s.live_deployment_id
WHERE s.kind = 'static_site' AND dep.status = 'live' AND dep.release_path IS NOT NULL AND dep.release_path != '';
