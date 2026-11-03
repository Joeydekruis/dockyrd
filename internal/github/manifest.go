package github

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

const defaultAppName = "Dockyard"

// ManifestParams configures the GitHub App manifest redirect flow.
type ManifestParams struct {
	PublicBaseURL string
	RedirectURL   string
	SetupURL      string
	WebhookURL    string
	AppName       string
}

// CreateManifestURL returns the GitHub settings URL to create an app from manifest.
func CreateManifestURL(p ManifestParams) (string, error) {
	base := strings.TrimRight(strings.TrimSpace(p.PublicBaseURL), "/")
	if base == "" {
		base = strings.TrimRight(os.Getenv("DOCKYARD_PUBLIC_URL"), "/")
	}
	if base == "" {
		return "", fmt.Errorf("public base URL required for GitHub manifest")
	}
	if p.RedirectURL == "" {
		p.RedirectURL = base + "/api/v1/github/app/callback"
	}
	if p.WebhookURL == "" {
		p.WebhookURL = base + "/api/v1/hooks/github"
	}
	if p.SetupURL == "" {
		p.SetupURL = base + "/settings/instance"
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
		"redirect_url":          p.RedirectURL,
		"setup_url":             p.SetupURL,
		"public":                true,
		"default_permissions": map[string]any{
			"contents":        "read",
			"metadata":        "read",
			"pull_requests":   "read",
		},
		"default_events": []string{"push"},
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("manifest", string(raw))
	return "https://github.com/settings/apps/new?" + q.Encode(), nil
}
