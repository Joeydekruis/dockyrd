package server

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Most httptest flows use owner login without TOTP; policy test opts in.
	if os.Getenv("dockyrd_REQUIRE_OWNER_2FA") == "" {
		_ = os.Setenv("dockyrd_REQUIRE_OWNER_2FA", "0")
	}
	os.Exit(m.Run())
}
