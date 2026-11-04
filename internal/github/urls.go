package github

import "fmt"

// InstallURL opens GitHub's flow to attach repositories to an existing app.
func InstallURL(slug string) string {
	slug = stringsTrim(slug)
	if slug == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/apps/%s/installations/new", slug)
}

// ManageInstallationURL opens installation settings on GitHub.
func ManageInstallationURL(installationID int64) string {
	if installationID <= 0 {
		return ""
	}
	return fmt.Sprintf("https://github.com/settings/installations/%d", installationID)
}
