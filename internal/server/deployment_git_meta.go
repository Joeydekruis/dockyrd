package server

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/github"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
)

func (d apiDeps) gitMetaForServiceDeploy(ctx context.Context, serviceID string) jobs.GitMeta {
	raw, err := github.LoadServiceConfig(ctx, d.q, serviceID)
	if err != nil {
		return jobs.GitMeta{}
	}
	cfg, err := appconfig.Parse(raw)
	if err != nil {
		return jobs.GitMeta{}
	}
	src := cfg.Source
	branch := src.BranchOrDefault()
	meta := jobs.GitMeta{Branch: branch}
	if strings.ToLower(strings.TrimSpace(src.Type)) != "github" {
		return meta
	}
	repo := src.GitHubRepoFullName()
	if repo == "" {
		return meta
	}
	creds, err := github.LoadCredentials(ctx, d.q, d.cfg.SecretKeyPath())
	if err != nil {
		return meta
	}
	token, err := creds.InstallationAccessToken(ctx)
	if err != nil {
		return meta
	}
	head, err := github.HeadCommitForBranch(ctx, token, repo, branch)
	if err != nil {
		return meta
	}
	return jobs.GitMeta{
		CommitSHA:     head.SHA,
		CommitMessage: head.Message,
		CommitAuthor:  head.Author,
		Branch:        branch,
	}
}

func gitMetaFromDeploymentRow(row sqlc.GetDeploymentByIDRow) jobs.GitMeta {
	return jobs.GitMeta{
		CommitSHA:     nullStringValue(row.CommitSha),
		CommitMessage: nullStringValue(row.CommitMessage),
		CommitAuthor:  nullStringValue(row.CommitAuthor),
		Branch:        nullStringValue(row.Branch),
	}
}

func mergeGitMeta(primary, fallback jobs.GitMeta) jobs.GitMeta {
	out := primary
	if out.CommitSHA == "" {
		out.CommitSHA = fallback.CommitSHA
	}
	if out.CommitMessage == "" {
		out.CommitMessage = fallback.CommitMessage
	}
	if out.CommitAuthor == "" {
		out.CommitAuthor = fallback.CommitAuthor
	}
	if out.Branch == "" {
		out.Branch = fallback.Branch
	}
	return out
}

func nullStringValue(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return strings.TrimSpace(v.String)
}
