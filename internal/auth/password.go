package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/alexedwards/argon2id"
)

const (
	argonMemory      = 64 * 1024
	argonIterations  = 3
	argonParallelism = 2
	argonKeyLength   = 32
)

// HashPassword returns an argon2id encoded hash for storage.
func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, &argon2id.Params{
		Memory:      argonMemory,
		Iterations:  argonIterations,
		Parallelism: argonParallelism,
		SaltLength:  16,
		KeyLength:   argonKeyLength,
	})
}

// VerifyPassword compares a password with an argon2id hash.
func VerifyPassword(password, encodedHash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, encodedHash)
}

// NewSessionToken returns a URL-safe opaque token and its sha256 hash bytes.
func NewSessionToken() (plain string, hash []byte, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, fmt.Errorf("rand: %w", err)
	}
	plain = hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(plain))
	return plain, sum[:], nil
}

// HashSessionToken hashes the cookie value for DB lookup.
func HashSessionToken(plain string) []byte {
	sum := sha256.Sum256([]byte(plain))
	return sum[:]
}
