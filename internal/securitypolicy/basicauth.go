package securitypolicy

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// GenerateBasicAuthPassword returns a one-time plaintext password and bcrypt hash for Caddy.
func GenerateBasicAuthPassword(username string) (password string, hash string, err error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", "", fmt.Errorf("username required")
	}
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	password = base64.RawURLEncoding.EncodeToString(buf)
	digest, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return password, string(digest), nil
}
