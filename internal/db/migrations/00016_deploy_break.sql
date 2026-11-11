-- +goose Up
CREATE TABLE deployment_break_scores (
  id TEXT PRIMARY KEY,
  deployment_id TEXT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  score INTEGER NOT NULL CHECK (score >= 0),
  updated_at INTEGER NOT NULL,
  UNIQUE (deployment_id, user_id)
);

CREATE INDEX idx_deployment_break_scores_deployment ON deployment_break_scores (deployment_id);
CREATE INDEX idx_deployment_break_scores_score ON deployment_break_scores (score DESC);

-- +goose Down
DROP TABLE IF EXISTS deployment_break_scores;
