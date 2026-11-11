package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// HeadCommitMeta is the branch tip commit as returned by the GitHub REST API.
type HeadCommitMeta struct {
	SHA     string
	Message string
	Author  string
}

// HeadCommitForBranch loads the latest commit on branch for owner/repo (e.g. acme/api).
func HeadCommitForBranch(ctx context.Context, token, ownerRepo, branch string) (HeadCommitMeta, error) {
	ownerRepo = strings.Trim(strings.TrimSpace(ownerRepo), "/")
	branch = strings.TrimSpace(branch)
	if ownerRepo == "" || branch == "" {
		return HeadCommitMeta{}, fmt.Errorf("repo and branch required")
	}
	parts := strings.SplitN(ownerRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return HeadCommitMeta{}, fmt.Errorf("invalid repo %q", ownerRepo)
	}
	encBranch := url.PathEscape(branch)
	apiURL := fmt.Sprintf("%s/repos/%s/%s/branches/%s", githubAPI, parts[0], parts[1], encBranch)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return HeadCommitMeta{}, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return HeadCommitMeta{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return HeadCommitMeta{}, fmt.Errorf("branch commit %d: %s", resp.StatusCode, string(body))
	}
	var out struct {
		Commit struct {
			SHA    string `json:"sha"`
			Commit struct {
				Message string `json:"message"`
				Author  struct {
					Name string `json:"name"`
				} `json:"author"`
			} `json:"commit"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return HeadCommitMeta{}, err
	}
	meta := HeadCommitMeta{
		SHA:     strings.TrimSpace(out.Commit.SHA),
		Message: strings.TrimSpace(out.Commit.Commit.Message),
		Author:  strings.TrimSpace(out.Commit.Commit.Author.Name),
	}
	if meta.SHA == "" {
		return HeadCommitMeta{}, fmt.Errorf("empty commit sha")
	}
	return meta, nil
}
