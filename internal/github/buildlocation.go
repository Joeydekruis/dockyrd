package github

import (
	"context"
	"encoding/json"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// ServiceUsesGitHubActions reports buildLocation=github_actions on an app service.
func ServiceUsesGitHubActions(raw json.RawMessage) bool {
	cfg, err := appconfig.Parse(raw)
	if err != nil {
		return false
	}
	return cfg.UsesGitHubActionsBuild()
}

// RepoFromServiceConfig returns owner/repo from git source.
func RepoFromServiceConfig(raw json.RawMessage) string {
	cfg, err := appconfig.Parse(raw)
	if err != nil {
		return ""
	}
	return cfg.Source.Repo
}

// LoadServiceConfig returns raw config for a service id.
func LoadServiceConfig(ctx context.Context, q *sqlc.Queries, serviceID string) (json.RawMessage, error) {
	row, err := q.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	return appconfig.RawConfig(row.Config), nil
}