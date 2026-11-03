package build

import (
	"errors"
	"os"
	"testing"
)

func TestCheckDiskGuardSkipsWhenUnknown(t *testing.T) {
	if err := CheckDiskGuard("", 10); err != nil {
		t.Fatal(err)
	}
}

func TestCheckDiskGuardOnRealDir(t *testing.T) {
	dir := t.TempDir()
	if err := CheckDiskGuard(dir, 10); err != nil {
		if errors.Is(err, ErrDiskLow) {
			t.Skip("disk too full in CI")
		}
		t.Fatal(err)
	}
	_ = os.WriteFile(dir+"/x", []byte("ok"), 0o644)
}
