-- +goose Up
CREATE TABLE addons (
  id TEXT PRIMARY KEY,
  state TEXT NOT NULL CHECK (state IN (
    'not_installed', 'installing', 'installed', 'failed', 'uninstalling'
  )),
  version TEXT,
  config_json TEXT,
  installed_at INTEGER,
  last_error TEXT
);

CREATE TABLE security_check_cache (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  ran_at INTEGER NOT NULL,
  results_json TEXT NOT NULL
);

INSERT INTO addons (id, state) VALUES
  ('firewall', 'not_installed'),
  ('unattended-updates', 'not_installed'),
  ('crowdsec', 'not_installed'),
  ('proxy-waf', 'not_installed'),
  ('proxy-rate-limit', 'not_installed');

-- +goose Down
DROP TABLE IF EXISTS security_check_cache;
DROP TABLE IF EXISTS addons;
