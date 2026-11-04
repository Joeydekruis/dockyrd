package staticsite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveStaticWorkDir_repoRoot(t *testing.T) {
	clone := t.TempDir()
	got, err := resolveStaticWorkDir(clone, "/", "main")
	if err != nil || got != clone {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestResolveStaticWorkDir_subdirectory(t *testing.T) {
	clone := t.TempDir()
	site := filepath.Join(clone, "site")
	if err := os.Mkdir(site, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := resolveStaticWorkDir(clone, "site", "main")
	if err != nil || got != site {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestResolveStaticWorkDir_missing(t *testing.T) {
	clone := t.TempDir()
	_, err := resolveStaticWorkDir(clone, "site", "development")
	if err == nil {
		t.Fatal("expected error")
	}
}
