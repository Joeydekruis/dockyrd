package crypto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureMasterKeyCreatesOnce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.key")

	created, err := EnsureMasterKey(path)
	if err != nil || !created {
		t.Fatalf("first EnsureMasterKey: created=%v err=%v", created, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o", info.Mode().Perm())
	}

	created, err = EnsureMasterKey(path)
	if err != nil || created {
		t.Fatalf("second EnsureMasterKey: created=%v err=%v", created, err)
	}
}
