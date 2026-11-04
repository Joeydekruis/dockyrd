package github

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func TestBuildManifestLaunch(t *testing.T) {
	launch, err := BuildManifestLaunch(ManifestParams{
		PublicBaseURL: "https://deploy.example.com",
		AppName:       "dockyrd-test",
	}, "state-token-1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(launch.Action, "https://github.com/settings/apps/new?state=") {
		t.Fatalf("action = %q", launch.Action)
	}
	if launch.State != "state-token-1" {
		t.Fatalf("state = %q", launch.State)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(launch.Manifest), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed["name"] != "dockyrd-test" {
		t.Fatalf("name = %v", parsed["name"])
	}
	if parsed["redirect_url"] != "https://deploy.example.com/api/v1/github/app/callback" {
		t.Fatalf("redirect_url = %v", parsed["redirect_url"])
	}
	hook, ok := parsed["hook_attributes"].(map[string]any)
	if !ok || hook["url"] != "https://deploy.example.com/api/v1/hooks/github" {
		t.Fatalf("hook_attributes = %v", parsed["hook_attributes"])
	}
}

func TestBuildManifestLaunchOrganization(t *testing.T) {
	launch, err := BuildManifestLaunch(ManifestParams{
		PublicBaseURL: "https://deploy.example.com",
		Organization:  "acme-corp",
	}, "s")
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := "https://github.com/organizations/acme-corp/settings/apps/new"
	if !strings.HasPrefix(launch.Action, wantPrefix) {
		t.Fatalf("action = %q", launch.Action)
	}
	parsed, err := url.Parse(launch.Action)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Query().Get("state") != "s" {
		t.Fatalf("state query = %q", parsed.Query().Get("state"))
	}
}
