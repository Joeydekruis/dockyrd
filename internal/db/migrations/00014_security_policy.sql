-- +goose Up
CREATE TABLE security_policy (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  default_profile TEXT NOT NULL CHECK (default_profile IN ('baseline', 'strict')),
  lock INTEGER NOT NULL DEFAULT 0 CHECK (lock IN (0, 1)),
  headers_json TEXT NOT NULL,
  rl_default_rpm INTEGER,
  rl_login_rpm INTEGER,
  cdn_mode TEXT NOT NULL DEFAULT 'direct' CHECK (cdn_mode IN ('direct', 'cloudflare', 'custom')),
  cdn_ranges_json TEXT NOT NULL DEFAULT '[]',
  cdn_header TEXT NOT NULL DEFAULT '',
  default_mem_mb INTEGER NOT NULL DEFAULT 512,
  default_cpu REAL NOT NULL DEFAULT 1.0,
  default_pids INTEGER NOT NULL DEFAULT 512
);

CREATE TABLE service_security (
  service_id TEXT PRIMARY KEY REFERENCES services(id) ON DELETE CASCADE,
  profile TEXT NOT NULL DEFAULT 'inherit' CHECK (profile IN ('inherit', 'baseline', 'strict', 'custom')),
  lock_exempt INTEGER NOT NULL DEFAULT 0 CHECK (lock_exempt IN (0, 1)),
  csp TEXT,
  csp_report_only INTEGER NOT NULL DEFAULT 0 CHECK (csp_report_only IN (0, 1)),
  basic_user TEXT,
  basic_hash TEXT,
  ip_allowlist_json TEXT NOT NULL DEFAULT '[]',
  maintenance INTEGER NOT NULL DEFAULT 0 CHECK (maintenance IN (0, 1)),
  rl_mode TEXT NOT NULL DEFAULT 'inherit' CHECK (rl_mode IN ('inherit', 'instance', 'custom', 'off')),
  rl_rpm INTEGER,
  rl_login_rpm INTEGER,
  waf_mode TEXT NOT NULL DEFAULT 'inherit' CHECK (waf_mode IN ('inherit', 'off', 'detect', 'block')),
  mem_mb INTEGER,
  cpu REAL,
  pids INTEGER
);

INSERT INTO security_policy (
  id, default_profile, lock, headers_json, cdn_mode, cdn_ranges_json, cdn_header,
  default_mem_mb, default_cpu, default_pids
) VALUES (
  1,
  'baseline',
  0,
  '{"Strict-Transport-Security":{"enabled":true,"value":"max-age=15552000"},"X-Content-Type-Options":{"enabled":true,"value":"nosniff"},"Referrer-Policy":{"enabled":true,"value":"strict-origin-when-cross-origin"},"X-Frame-Options":{"enabled":true,"value":"SAMEORIGIN"}}',
  'direct',
  '[]',
  '',
  512,
  1.0,
  512
);

-- +goose Down
DROP TABLE IF EXISTS service_security;
DROP TABLE IF EXISTS security_policy;
