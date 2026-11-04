package staticsite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPublishRelease_relativeCurrentSymlink(t *testing.T) {
	sitesRoot := t.TempDir()
	output := t.TempDir()
	if err := os.WriteFile(filepath.Join(output, "index.html"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := PublishRelease(output, sitesRoot, "demo", "dep1")
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(sitesRoot, "demo", "current")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(target) {
		t.Fatalf("expected relative symlink target, got %q", target)
	}
	want := filepath.Join("releases", "dep1")
	if target != want {
		t.Fatalf("got %q want %q", target, want)
	}
	if _, err := os.Stat(filepath.Join(link, "index.html")); err != nil {
		t.Fatalf("index via symlink: %v", err)
	}
}

func TestEnsureRelativeCurrentLink_repairsAbsolute(t *testing.T) {
	sitesRoot := t.TempDir()
	paths := PathsForService(sitesRoot, "demo")
	if err := os.MkdirAll(paths.ReleasesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	release := filepath.Join(paths.ReleasesDir, "dep1")
	if err := os.Mkdir(release, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(release, paths.CurrentLink); err != nil {
		t.Fatal(err)
	}
	if err := EnsureRelativeCurrentLink(sitesRoot, "demo"); err != nil {
		t.Fatal(err)
	}
	target, err := os.Readlink(paths.CurrentLink)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(target) {
		t.Fatalf("still absolute: %q", target)
	}
}
