package preview

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/github"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
)

type parentMatch struct {
	Repo   string
	Branch string
}

// HandlePullRequest creates, updates, or removes preview child services for a PR webhook.
func HandlePullRequest(ctx context.Context, q *sqlc.Queries, opts SyncOptions, ev github.PullRequestEvent) error {
	repo := strings.TrimSpace(ev.Repository.FullName)
	baseBranch := github.BranchFromRef(ev.PullRequest.Base.Ref)
	switch ev.Action {
	case "closed":
		return deletePreviewForPR(ctx, q, repo, baseBranch, ev.Number)
	case "opened", "synchronize", "reopened":
		return upsertPreviewDeploy(ctx, q, opts, ev, repo, baseBranch)
	default:
		return nil
	}
}

func upsertPreviewDeploy(ctx context.Context, q *sqlc.Queries, opts SyncOptions, ev github.PullRequestEvent, repo, baseBranch string) error {
	parents, err := listPreviewParents(ctx, q, repo, baseBranch)
	if err != nil {
		return err
	}
	headSHA := strings.TrimSpace(ev.PullRequest.Head.Sha)
	headBranch := github.HeadBranchFromPR(ev.PullRequest.Head.Ref)
	now := time.Now().Unix()
	meta := jobs.GitMeta{
		CommitSHA: headSHA,
		Branch:    headBranch,
	}
	for _, parentID := range parents {
		parent, err := q.GetServiceByID(ctx, parentID)
		if err != nil {
			return err
		}
		childID := ""
		existing, err := q.GetServiceByParentAndPR(ctx, sqlc.GetServiceByParentAndPRParams{
			ParentServiceID: sql.NullString{String: parentID, Valid: true},
			PreviewPr:       sql.NullInt64{Int64: int64(ev.Number), Valid: true},
		})
		if err == nil {
			childID = existing.ID
			if err := q.UpdatePreviewPushMeta(ctx, sqlc.UpdatePreviewPushMetaParams{
				PreviewHeadSha:    sql.NullString{String: headSHA, Valid: headSHA != ""},
				PreviewLastPushAt: sql.NullInt64{Int64: now, Valid: true},
				UpdatedAt:         sql.NullInt64{Int64: now, Valid: true},
				ID:                childID,
			}); err != nil {
				return err
			}
		} else if errors.Is(err, sql.ErrNoRows) {
			childID, err = createPreviewChild(ctx, q, opts, parent, ev.Number, headSHA, now)
			if err != nil {
				return err
			}
		} else {
			return err
		}
		if _, err := jobs.EnqueueServiceDeploy(ctx, q, childID, "preview", "", meta); err != nil {
			return err
		}
	}
	return nil
}

func createPreviewChild(
	ctx context.Context,
	q *sqlc.Queries,
	opts SyncOptions,
	parent sqlc.Service,
	pr int,
	headSHA string,
	now int64,
) (string, error) {
	baseSlug := parent.Slug + "-pr-" + fmt.Sprintf("%d", pr)
	slug := baseSlug
	for i := 0; i < 5; i++ {
		_, err := q.ServiceSlugExistsInEnvironment(ctx, sqlc.ServiceSlugExistsInEnvironmentParams{
			EnvironmentID: parent.EnvironmentID,
			Slug:          slug,
		})
		if errors.Is(err, sql.ErrNoRows) {
			break
		}
		if err != nil {
			return "", err
		}
		slug = baseSlug + "-" + ulid.Make().String()[22:]
	}
	id := ulid.Make().String()
	name := fmt.Sprintf("PR #%d preview", pr)
	if err := q.CreatePreviewChildService(ctx, sqlc.CreatePreviewChildServiceParams{
		ID:                id,
		EnvironmentID:     parent.EnvironmentID,
		Kind:              parent.Kind,
		Name:              name,
		Slug:              slug,
		DesiredState:      parent.DesiredState,
		Config:            parent.Config,
		CreatedAt:         sql.NullInt64{Int64: now, Valid: true},
		UpdatedAt:         sql.NullInt64{Int64: now, Valid: true},
		ParentServiceID:   sql.NullString{String: parent.ID, Valid: true},
		PreviewPr:         sql.NullInt64{Int64: int64(pr), Valid: true},
		PreviewHeadSha:    sql.NullString{String: headSHA, Valid: headSHA != ""},
		PreviewLastPushAt: sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		return "", err
	}
	if opts.SecretKeyPath != "" {
		_ = CopyVariables(ctx, q, opts.SecretKeyPath, parent.ID, id)
	}
	_ = EnsurePlatformDomain(ctx, q, id, slug, opts.PublicIPv4)
	return id, nil
}

func deletePreviewForPR(ctx context.Context, q *sqlc.Queries, repo, baseBranch string, pr int) error {
	parents, err := listPreviewParents(ctx, q, repo, baseBranch)
	if err != nil {
		return err
	}
	for _, parentID := range parents {
		child, err := q.GetServiceByParentAndPR(ctx, sqlc.GetServiceByParentAndPRParams{
			ParentServiceID: sql.NullString{String: parentID, Valid: true},
			PreviewPr:       sql.NullInt64{Int64: int64(pr), Valid: true},
		})
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		if err := q.DeleteService(ctx, child.ID); err != nil {
			return err
		}
	}
	return nil
}

func listPreviewParents(ctx context.Context, q *sqlc.Queries, repoFullName, branch string) ([]string, error) {
	rows, err := q.ListDeployableServices(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0)
	for _, row := range rows {
		if !EnabledRaw(row.Config) {
			continue
		}
		match, ok := deploySourceMatch(row.Config, row.Kind)
		if !ok {
			continue
		}
		if strings.TrimSpace(match.Repo) != strings.TrimSpace(repoFullName) {
			continue
		}
		wantBranch := strings.TrimSpace(match.Branch)
		if wantBranch == "" {
			wantBranch = "main"
		}
		if wantBranch != strings.TrimSpace(branch) {
			continue
		}
		out = append(out, row.ID)
	}
	return out, nil
}

func deploySourceMatch(raw any, kind string) (parentMatch, bool) {
	if kind != "app" && kind != "static_site" {
		return parentMatch{}, false
	}
	cfg, err := json.Marshal(raw)
	if err != nil {
		return parentMatch{}, false
	}
	var m struct {
		Source struct {
			Type   string `json:"type"`
			Repo   string `json:"repo"`
			Branch string `json:"branch"`
		} `json:"source"`
	}
	if err := json.Unmarshal(cfg, &m); err != nil {
		return parentMatch{}, false
	}
	if strings.ToLower(strings.TrimSpace(m.Source.Type)) != "github" {
		return parentMatch{}, false
	}
	if strings.TrimSpace(m.Source.Repo) == "" {
		return parentMatch{}, false
	}
	return parentMatch{Repo: m.Source.Repo, Branch: m.Source.Branch}, true
}
