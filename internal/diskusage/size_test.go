package diskusage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirSizeAndFileSize(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(dir, "nested")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), make([]byte, 100), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := FileSize(filepath.Join(dir, "a.txt")); got != 5 {
		t.Fatalf("file size=%d", got)
	}
	if got := DirSize(dir); got != 105 {
		t.Fatalf("dir size=%d", got)
	}
}

func TestDataDirCategories(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "control.db"), make([]byte, 50), 0o644); err != nil {
		t.Fatal(err)
	}
	caddy := filepath.Join(dir, "caddy")
	if err := os.MkdirAll(caddy, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(caddy, "access.log"), make([]byte, 200), 0o644); err != nil {
		t.Fatal(err)
	}
	cats := dataDirCategories(dir)
	if len(cats) != 2 {
		t.Fatalf("categories=%d", len(cats))
	}
	if cats[0].ID != "caddy" || cats[0].Bytes != 200 {
		t.Fatalf("caddy cat=%+v", cats[0])
	}
}
