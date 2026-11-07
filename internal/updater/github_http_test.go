package updater

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchLatestSendsGitHubToken(t *testing.T) {
	t.Setenv("dockyrd_GITHUB_TOKEN", "ghp_test_token")
	t.Setenv("GITHUB_TOKEN", "ignored")

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v0.0.1"}`))
	}))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	applyGitHubAPIHeaders(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if gotAuth != "Bearer ghp_test_token" {
		t.Fatalf("Authorization=%q want Bearer ghp_test_token", gotAuth)
	}
}

func TestGitHubTokenPrefersdockyrdEnv(t *testing.T) {
	t.Setenv("dockyrd_GITHUB_TOKEN", "dockyrd_pat")
	t.Setenv("GITHUB_TOKEN", "generic")
	if got := GitHubToken(); got != "dockyrd_pat" {
		t.Fatalf("GitHubToken()=%q", got)
	}
}

func TestGitHubTokenFallback(t *testing.T) {
	t.Setenv("dockyrd_GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "generic")
	if got := GitHubToken(); got != "generic" {
		t.Fatalf("GitHubToken()=%q", got)
	}
}

func TestGithubRateLimitHint(t *testing.T) {
	h := githubRateLimitHint(http.StatusForbidden, `rate limit exceeded`)
	if h == "" {
		t.Fatal("expected hint")
	}
	if !strings.Contains(h, "dockyrd_GITHUB_TOKEN") {
		t.Fatalf("hint=%q", h)
	}
}
