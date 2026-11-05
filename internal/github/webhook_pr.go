package github

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PullRequestEvent is the subset of a GitHub pull_request webhook we care about.
type PullRequestEvent struct {
	Action string `json:"action"`
	Number int    `json:"number"`
	PullRequest struct {
		Head struct {
			Sha string `json:"sha"`
			Ref string `json:"ref"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

// ParsePullRequestEvent decodes a pull_request delivery body.
func ParsePullRequestEvent(body []byte) (PullRequestEvent, error) {
	var ev PullRequestEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		return PullRequestEvent{}, err
	}
	if ev.Repository.FullName == "" {
		return PullRequestEvent{}, fmt.Errorf("missing repository.full_name")
	}
	if ev.Number == 0 {
		return PullRequestEvent{}, fmt.Errorf("missing pull request number")
	}
	ev.Action = strings.TrimSpace(strings.ToLower(ev.Action))
	return ev, nil
}

// HeadBranchFromPR returns the head branch name (refs/heads/foo → foo).
func HeadBranchFromPR(ref string) string {
	return BranchFromRef(ref)
}
