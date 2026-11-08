package build

import (
	"testing"
)

func TestRailpackAssetTarget(t *testing.T) {
	target, err := railpackAssetTarget()
	if err != nil {
		t.Fatalf("railpackAssetTarget: %v", err)
	}
	switch target {
	case "x86_64-unknown-linux-musl", "arm64-unknown-linux-musl",
		"x86_64-apple-darwin", "arm64-apple-darwin":
	default:
		t.Fatalf("unexpected target %q on this host", target)
	}
}

func TestDefaultRailpackFrontendMatchesVersion(t *testing.T) {
	want := "ghcr.io/railwayapp/railpack-frontend:v" + DefaultRailpackVersion
	if DefaultRailpackFrontend != want {
		t.Fatalf("frontend %q want %q", DefaultRailpackFrontend, want)
	}
}
