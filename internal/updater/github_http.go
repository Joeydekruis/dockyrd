package updater

import (
	"net/http"
	"os"
	"strings"
)

const githubAPIVersion = "application/vnd.github+json"

// GitHubToken returns a PAT for GitHub API calls (releases check). Empty if unset.
func GitHubToken() string {
	if t := strings.TrimSpace(os.Getenv("DOCKYARD_GITHUB_TOKEN")); t != "" {
		return t
	}
	return strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
}

func applyGitHubAPIHeaders(req *http.Request) {
	req.Header.Set("Accept", githubAPIVersion)
	req.Header.Set("User-Agent", "dockyard-updater")
	if t := GitHubToken(); t != "" {
		req.Header.Set("Authorization", "Bearer "+t)
	}
}

func githubRateLimitHint(statusCode int, body string) string {
	if statusCode != http.StatusForbidden {
		return ""
	}
	if !strings.Contains(body, "rate limit") {
		return ""
	}
	return " Set DOCKYARD_GITHUB_TOKEN (or GITHUB_TOKEN) in /etc/dockyard/dockyard.env for a higher GitHub API limit."
}
