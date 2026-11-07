package updater

import "testing"

func TestReleaseAssetPrefixes(t *testing.T) {
	t.Parallel()
	got := releaseAssetPrefixes("v0.7.2", "linux", "arm64")
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0] != "dockyrd_0.7.2_linux_arm64" {
		t.Fatalf("canonical=%q", got[0])
	}
	if got[1] != "dockyard_0.7.2_linux_arm64" {
		t.Fatalf("legacy=%q", got[1])
	}
}
