-- name: GetSecurityPolicy :one
SELECT
  id, default_profile, lock, headers_json,
  rl_default_rpm, rl_login_rpm,
  cdn_mode, cdn_ranges_json, cdn_header,
  default_mem_mb, default_cpu, default_pids
FROM security_policy
WHERE id = 1;

-- name: UpdateSecurityPolicy :exec
UPDATE security_policy SET
  default_profile = ?,
  lock = ?,
  headers_json = ?,
  rl_default_rpm = ?,
  rl_login_rpm = ?,
  cdn_mode = ?,
  cdn_ranges_json = ?,
  cdn_header = ?,
  default_mem_mb = ?,
  default_cpu = ?,
  default_pids = ?
WHERE id = 1;

-- name: GetServiceSecurity :one
SELECT
  service_id, profile, lock_exempt, csp, csp_report_only,
  basic_user, basic_hash, ip_allowlist_json, maintenance,
  rl_mode, rl_rpm, rl_login_rpm, waf_mode, mem_mb, cpu, pids
FROM service_security
WHERE service_id = ?;

-- name: InsertServiceSecurity :exec
INSERT INTO service_security (
  service_id, profile, lock_exempt, csp, csp_report_only,
  basic_user, basic_hash, ip_allowlist_json, maintenance,
  rl_mode, rl_rpm, rl_login_rpm, waf_mode, mem_mb, cpu, pids
) VALUES (
  ?, ?, ?, ?, ?,
  ?, ?, ?, ?,
  ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateServiceSecurity :exec
UPDATE service_security SET
  profile = ?,
  lock_exempt = ?,
  csp = ?,
  csp_report_only = ?,
  basic_user = ?,
  basic_hash = ?,
  ip_allowlist_json = ?,
  maintenance = ?,
  rl_mode = ?,
  rl_rpm = ?,
  rl_login_rpm = ?,
  waf_mode = ?,
  mem_mb = ?,
  cpu = ?,
  pids = ?
WHERE service_id = ?;

-- name: ClearServiceBasicAuth :exec
UPDATE service_security SET basic_user = NULL, basic_hash = NULL
WHERE service_id = ?;