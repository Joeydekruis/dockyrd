package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeadCommitForBranch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/app/branches/main" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"commit": map[string]any{
				"sha": "abc123def",
				"commit": map[string]any{
					"message": "Refresh app icons\n\nBody line",
					"author": map[string]string{
						"name": "Joey de Kruis",
					},
				},
			},
		})
	}))
	defer srv.Close()

	orig := githubAPI
	githubAPI = srv.URL
	defer func() { githubAPI = orig }()

	meta, err := HeadCommitForBranch(context.Background(), "tok", "acme/app", "main")
	if err != nil {
		t.Fatal(err)
	}
	if meta.SHA != "abc123def" {
		t.Fatalf("sha=%q", meta.SHA)
	}
	if meta.Message != "Refresh app icons\n\nBody line" {
		t.Fatalf("message=%q", meta.Message)
	}
	if meta.Author != "Joey de Kruis" {
		t.Fatalf("author=%q", meta.Author)
	}
}
