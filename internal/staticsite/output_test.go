package staticsite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectOutputDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := DetectOutputDir(dir, "")
	if err != nil || got != filepath.Join(dir, "dist") {
		t.Fatalf("got %q err=%v", got, err)
	}
}
