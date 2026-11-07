package github

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

const defaultAppName = "dockyrd"

// ManifestParams configures the GitHub App manifest redirect flow.
type ManifestParams struct {
	PublicBaseURL string
	RedirectURL   string
	SetupURL      string
	WebhookURL    string
	AppName       string
	Organization  string
}

// ManifestLaunch is submitted to GitHub as POST application/x-www-form-urlencoded.
// GitHub requires POST with a form field named manifest — a GET ?manifest= query does not prefill.
type ManifestLaunch struct {
	Action   string `json:"action"`
	State    string `json:"state"`
	Manifest string `json:"manifest"`
}

// BuildManifestLaunch prepares the GitHub App manifest registration handshake.
func BuildManifestLaunch(p ManifestParams, state string) (ManifestLaunch, error) {
	base := strings.TrimRight(strings.TrimSpace(p.PublicBaseURL), "/")
	if base == "" {
		base = strings.TrimRight(os.Getenv("dockyrd_PUBLIC_URL"), "/")
	}
	if base == "" {
		return ManifestLaunch{}, fmt.Errorf("public base URL required for GitHub manifest")
	}
	if p.RedirectURL == "" {
		p.RedirectURL = base + "/api/v1/github/app/callback"
	}
	if p.WebhookURL == "" {
		p.WebhookURL = base + "/api/v1/hooks/github"
	}
	if p.SetupURL == "" {
		p.SetupURL = base + "/settings/server/integrations"
	}
	name := strings.TrimSpace(p.AppName)
	if name == "" {
		name = defaultAppName
	}
	manifest := map[string]any{
		"name": name,
		"url":  base,
		"hook_attributes": map[string]any{
			"url": p.WebhookURL,
		},
		"redirect_url": p.RedirectURL,
		"setup_url":    p.SetupURL,
		"public":       true,
		"default_permissions": map[string]any{
			"contents":      "read",
			"metadata":      "read",
			"pull_requests": "read",
		},
		"default_events": []string{"push", "pull_request"},
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		return ManifestLaunch{}, err
	}
	action := "https://github.com/settings/apps/new"
	if org := strings.TrimSpace(p.Organization); org != "" {
		action = fmt.Sprintf("https://github.com/organizations/%s/settings/apps/new", url.PathEscape(org))
	}
	if strings.TrimSpace(state) != "" {
		action += "?state=" + url.QueryEscape(state)
	}
	return ManifestLaunch{
		Action:   action,
		State:    state,
		Manifest: string(raw),
	}, nil
}

// CreateManifestURL is deprecated: GitHub ignores manifest on GET. Use BuildManifestLaunch + POST form.
func CreateManifestURL(p ManifestParams) (string, error) {
	launch, err := BuildManifestLaunch(p, "")
	if err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("manifest", launch.Manifest)
	return launch.Action + "&" + q.Encode(), nil
}
