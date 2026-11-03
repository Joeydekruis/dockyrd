package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// ResolveClone returns clone URL and branch for git/github sources.
func ResolveClone(ctx context.Context, q *sqlc.Queries, keyPath string, source appconfig.Source) (cloneURL, branch string, err error) {
	branch = source.BranchOrDefault()
	switch strings.ToLower(strings.TrimSpace(source.Type)) {
	case "git":
		return strings.TrimSpace(source.URL), branch, nil
	case "github":
		repo := strings.TrimSpace(source.Repo)
		if repo == "" {
			return "", "", fmt.Errorf("github repo required")
		}
		creds, err := LoadCredentials(ctx, q, keyPath)
		if err != nil {
			return AuthenticatedCloneURL("", repo), branch, nil
		}
		token, err := creds.InstallationAccessToken(ctx)
		if err != nil {
			return AuthenticatedCloneURL("", repo), branch, nil
		}
		return AuthenticatedCloneURL(token, repo), branch, nil
	default:
		return "", "", fmt.Errorf("unsupported source type %q", source.Type)
	}
}
