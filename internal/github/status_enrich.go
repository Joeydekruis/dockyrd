package github

import (
	"context"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// EnrichStatus fills URLs and live GitHub details when credentials exist.
func EnrichStatus(ctx context.Context, q *sqlc.Queries, keyPath, publicURL string, base StatusResponse) (StatusResponse, error) {
	out := base
	out.PublicURL = stringsTrim(publicURL)
	out.AppConfigured = out.Linked
	if out.Slug != "" {
		out.InstallURL = InstallURL(out.Slug)
	}
	if out.InstallationID > 0 {
		out.ManageURL = ManageInstallationURL(out.InstallationID)
	}
	row, err := q.GetGitSource(ctx, defaultGitSourceID)
	if err != nil {
		return out, nil
	}
	out.WebhookActive = len(row.WebhookSecretEnc) > 0
	if !out.Linked || out.InstallationID == 0 {
		return out, nil
	}
	creds, err := LoadCredentials(ctx, q, keyPath)
	if err != nil {
		return out, nil
	}
	acc, err := creds.InstallationAccount(ctx)
	if err == nil && acc.Login != "" {
		out.InstalledOn = acc.Login
	}
	token, err := creds.InstallationAccessToken(ctx)
	if err != nil {
		return out, nil
	}
	repos, err := ListRepos(ctx, token, "")
	if err != nil {
		return out, nil
	}
	out.RepositoryCount = len(repos)
	out.InstallationReady = out.InstallationID > 0 && out.RepositoryCount > 0
	return out, nil
}
