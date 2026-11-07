-- +goose Up
CREATE TABLE host_actions (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  state TEXT NOT NULL CHECK (state IN (
    'planned', 'applied', 'verified', 'awaiting_confirm',
    'confirmed', 'reverted', 'failed', 'cancelled'
  )),
  plan_json TEXT,
  revert_json TEXT,
  actor_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
  applied_at INTEGER,
  confirm_deadline INTEGER,
  confirmed_at INTEGER,
  reverted_at INTEGER,
  last_error TEXT,
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

CREATE INDEX idx_host_actions_reconcile ON host_actions (state, confirm_deadline)
WHERE state = 'awaiting_confirm';

CREATE INDEX idx_host_actions_kind_created ON host_actions (kind, created_at DESC);
