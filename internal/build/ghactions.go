package build

import (
	"fmt"
	"strings"
)

// GitHubWorkflowInput configures generated workflow YAML.
type GitHubWorkflowInput struct {
	Branch      string
	Context     string
	Image       string
	DockyardURL string
	ServiceID   string
}

// RenderGitHubWorkflow returns dockyard.yml contents for copy into the repo.
func RenderGitHubWorkflow(in GitHubWorkflowInput) (string, error) {
	if strings.TrimSpace(in.Branch) == "" {
		in.Branch = "main"
	}
	if strings.TrimSpace(in.Context) == "" {
		in.Context = "."
	}
	base := strings.TrimRight(strings.TrimSpace(in.DockyardURL), "/")
	image := strings.TrimSpace(in.Image)
	return fmt.Sprintf(`name: Dockyard Deploy

on:
  push:
    branches:
      - %s

permissions:
  contents: read
  packages: write

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      - name: Log in to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: %s
          push: true
          tags: %s:${{ github.sha }}
      - name: Deploy to Dockyard
        run: |
          curl -sfS -X POST "%s/api/v1/hooks/deploy/%s" \
            -H "Authorization: Bearer ${{ secrets.DOCKYARD_DEPLOY_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d "{\"image\":\"%s:${{ github.sha }}\",\"commit_sha\":\"${{ github.sha }}\"}"
`, in.Branch, in.Context, image, base, in.ServiceID, image), nil
}

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
