package build

import (
	"strings"
	"testing"
)

func TestRenderGitHubWorkflow(t *testing.T) {
	out, err := RenderGitHubWorkflow(GitHubWorkflowInput{
		Branch:      "main",
		Context:     ".",
		Image:       "ghcr.io/acme/app",
		DockyrdURL: "https://dock.example.com",
		ServiceID:   "01TEST",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "hooks/deploy/01TEST") {
		t.Fatalf("missing hook path: %s", out)
	}
	if !strings.Contains(out, "dockyrd_DEPLOY_TOKEN") {
		t.Fatal("missing secret name")
	}
}

func TestImageMatchesRepo(t *testing.T) {
	if !ImageMatchesRepo("acme/app", "ghcr.io/acme/app:abc") {
		t.Fatal("expected match")
	}
	if ImageMatchesRepo("acme/app", "docker.io/other:abc") {
		t.Fatal("expected reject")
	}
}
