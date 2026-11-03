package auth

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestValidateTOTP_roundTrip(t *testing.T) {
	secret, _, err := NewTOTPSecret("user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	code, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if !ValidateTOTP(secret, code) {
		t.Fatal("expected valid code")
	}
	if ValidateTOTP(secret, "000000") {
		t.Fatal("expected invalid code")
	}
}

func TestRecoveryCodesUnique(t *testing.T) {
	plain, hashes, err := NewRecoveryCodes(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(plain) != 10 || len(hashes) != 10 {
		t.Fatalf("want 10 codes, got %d/%d", len(plain), len(hashes))
	}
	seen := map[string]struct{}{}
	for _, c := range plain {
		if len(c) != 8 {
			t.Fatalf("code length %d", len(c))
		}
		seen[c] = struct{}{}
	}
	if len(seen) != 10 {
		t.Fatal("duplicate recovery codes")
	}
}
