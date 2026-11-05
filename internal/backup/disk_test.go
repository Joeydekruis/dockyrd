package backup

import (
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/build"
)

func TestCheckDiskGuardWrapsBuildError(t *testing.T) {
	if err := CheckDiskGuard("", 10); err != nil {
		t.Fatal(err)
	}
	_ = build.ErrDiskLow
}
