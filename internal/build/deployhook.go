package build

import (
	"fmt"
	"strings"
)

// ExpectedGHCRImage returns ghcr.io/owner/repo for a GitHub repo slug.
func ExpectedGHCRImage(repo string) string {
	repo = strings.TrimSpace(repo)
	repo = strings.TrimPrefix(repo, "https://github.com/")
	repo = strings.TrimSuffix(repo, ".git")
	return fmt.Sprintf("ghcr.io/%s", strings.ToLower(repo))
}

// ImageMatchesRepo rejects deploy hooks pointing at foreign registries.
func ImageMatchesRepo(repo, image string) bool {
	wantPrefix := ExpectedGHCRImage(repo) + ":"
	image = strings.TrimSpace(image)
	return strings.HasPrefix(strings.ToLower(image), strings.ToLower(wantPrefix)) ||
		strings.HasPrefix(strings.ToLower(image), strings.ToLower(ExpectedGHCRImage(repo))+"@")
}