package build

import "testing"

func TestImageMatchesRepo(t *testing.T) {
	if !ImageMatchesRepo("acme/app", "ghcr.io/acme/app:abc") {
		t.Fatal("expected match")
	}
	if ImageMatchesRepo("acme/app", "docker.io/other:abc") {
		t.Fatal("expected no match")
	}
}