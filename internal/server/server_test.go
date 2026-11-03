package server

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Most httptest flows use owner login without TOTP; policy test opts in.
	if os.Getenv("DOCKYARD_REQUIRE_OWNER_2FA") == "" {
		_ = os.Setenv("DOCKYARD_REQUIRE_OWNER_2FA", "0")
	}
	os.Exit(m.Run())
}
