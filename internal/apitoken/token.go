package apitoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const PlainPrefix = "dyk_"

// Scope values for user-issued CLI tokens (stored in api_tokens.scope).
const (
	ScopeRead   = "read"
	ScopeDeploy = "deploy"
	ScopeWrite  = "write"
)

// Generate returns a one-time plaintext token (dyk_ prefix) and its sha256 hash bytes.
func Generate() (plain string, hash []byte, err error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", nil, err
	}
	plain = PlainPrefix + hex.EncodeToString(b[:])
	sum := sha256.Sum256([]byte(plain))
	return plain, sum[:], nil
}

// Hash computes sha256 of the presented token (Bearer prefix stripped).
func Hash(plain string) ([]byte, error) {
	plain = TrimBearer(plain)
	if plain == "" {
		return nil, fmt.Errorf("token required")
	}
	sum := sha256.Sum256([]byte(plain))
	return sum[:], nil
}

func TrimBearer(s string) string {
	const prefix = "Bearer "
	if len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix) {
		return strings.TrimSpace(s[len(prefix):])
	}
	return strings.TrimSpace(s)
}

// Prefix returns a short display prefix for a stored token (never the full secret).
func Prefix(plain string) string {
	plain = TrimBearer(plain)
	if len(plain) <= 12 {
		return plain
	}
	return plain[:12] + "…"
}

// ValidUserScope reports whether scope is read|deploy|write.
func ValidUserScope(scope string) bool {
	switch scope {
	case ScopeRead, ScopeDeploy, ScopeWrite:
		return true
	default:
		return false
	}
}

// ScopeRank orders scopes for comparison (higher = more privilege).
func ScopeRank(scope string) int {
	switch scope {
	case ScopeRead:
		return 1
	case ScopeDeploy:
		return 2
	case ScopeWrite:
		return 3
	default:
		return 0
	}
}

// ScopeAllows returns true when tokenScope meets minScope (read < deploy < write).
func ScopeAllows(tokenScope, minScope string) bool {
	return ScopeRank(tokenScope) >= ScopeRank(minScope)
}

// DeployScope returns the scope string for a service deploy hook.
func DeployScope(serviceID string) string {
	return "deploy:" + serviceID
}

// IsDeployHookScope is true for per-service GitHub Action hook tokens.
func IsDeployHookScope(scope string) bool {
	return strings.HasPrefix(scope, "deploy:")
}

// IsUserScope reports whether scope is a user CLI token (read|deploy|write).
func IsUserScope(scope string) bool {
	return ValidUserScope(scope)
}

// ValidateScope returns true when tokenScope meets need (read < deploy < write).
func ValidateScope(tokenScope, need string) bool {
	return ScopeAllows(tokenScope, need)
}
