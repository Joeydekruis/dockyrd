package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const totpIssuer = "dockyrd"

// NewTOTPSecret generates a base32 secret and otpauth URL for enrollment.
func NewTOTPSecret(accountEmail string) (secret string, otpauthURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: accountEmail,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

// ValidateTOTP checks a 6-digit code against the base32 secret.
func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// NewRecoveryCodes returns plaintext codes and their sha256 hashes for storage.
func NewRecoveryCodes(n int) (plain []string, hashes [][]byte, err error) {
	plain = make([]string, 0, n)
	hashes = make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		raw := make([]byte, 5)
		if _, err := rand.Read(raw); err != nil {
			return nil, nil, fmt.Errorf("rand: %w", err)
		}
		code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)[:8]
		sum := sha256.Sum256([]byte(code))
		plain = append(plain, code)
		hashes = append(hashes, sum[:])
	}
	return plain, hashes, nil
}

// HashRecoveryCode hashes a user-entered recovery code for lookup.
func HashRecoveryCode(code string) []byte {
	sum := sha256.Sum256([]byte(code))
	return sum[:]
}

// FormatRecoveryCodesForDisplay joins codes for one-time display.
func FormatRecoveryCodesForDisplay(codes []string) string {
	out := ""
	for i, c := range codes {
		if i > 0 {
			out += "\n"
		}
		out += c
	}
	return out
}
