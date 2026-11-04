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

func TestDetectOutputDir_plainHTMLDot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!doctype html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := DetectOutputDir(dir, ".")
	if err != nil || got != dir {
		t.Fatalf("got %q err=%v", got, err)
	}
}
