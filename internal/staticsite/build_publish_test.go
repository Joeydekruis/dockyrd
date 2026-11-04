package staticsite

import (
	"os"
	"path/filepath"
	"testing"
)

// PublishRelease must run while clone output still exists; deploy cleans src after copy.
func TestPublishRelease_afterDetectOutputDir(t *testing.T) {
	clone := t.TempDir()
	site := filepath.Join(clone, "site")
	if err := os.Mkdir(site, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(site, "index.html"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	workDir, err := resolveStaticWorkDir(clone, "site", "main")
	if err != nil {
		t.Fatal(err)
	}
	outputDir, err := DetectOutputDir(workDir, ".")
	if err != nil {
		t.Fatal(err)
	}
	sitesRoot := t.TempDir()
	_, err = PublishRelease(outputDir, sitesRoot, "demo", "dep1")
	if err != nil {
		t.Fatalf("PublishRelease: %v", err)
	}
	current := filepath.Join(sitesRoot, "demo", "current", "index.html")
	if _, err := os.Stat(current); err != nil {
		t.Fatalf("expected published index at %s: %v", current, err)
	}
}
