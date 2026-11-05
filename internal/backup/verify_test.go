package backup

import (
	"os"
	"testing"
)

func TestCheckDiskGuardOnTempDir(t *testing.T) {
	dir := t.TempDir()
	if err := CheckDiskGuard(dir, 10); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(dir+"/x", []byte("ok"), 0o644)
}

func TestDigestFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/f.bin"
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	sum, err := DigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(sum) != 64 {
		t.Fatalf("sha256 hex len %d", len(sum))
	}
}
