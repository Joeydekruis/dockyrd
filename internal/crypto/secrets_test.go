package crypto

import (
	"path/filepath"
	"testing"
)

func TestEncryptRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.key")
	if _, err := EnsureMasterKey(path); err != nil {
		t.Fatal(err)
	}
	key, err := LoadMasterKey(path)
	if err != nil {
		t.Fatal(err)
	}
	plain := []byte("DATABASE_URL=postgres://local")
	enc, err := Encrypt(key, plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decrypt(key, enc)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(plain) {
		t.Fatalf("got %q", got)
	}
}
