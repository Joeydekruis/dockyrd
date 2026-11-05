package apitoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Generate returns a one-time plaintext token and its sha256 hash bytes.
func Generate() (plain string, hash []byte, err error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", nil, err
	}
	plain = hex.EncodeToString(b[:])
	sum := sha256.Sum256([]byte(plain))
	return plain, sum[:], nil
}

// Hash computes sha256 of the presented token.
func Hash(plain string) ([]byte, error) {
	plain = trimToken(plain)
	if plain == "" {
		return nil, fmt.Errorf("token required")
	}
	sum := sha256.Sum256([]byte(plain))
	return sum[:], nil
}

func trimToken(s string) string {
	const prefix = "Bearer "
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

// DeployScope returns the scope string for a service deploy hook.
func DeployScope(serviceID string) string {
	return "deploy:" + serviceID
}
