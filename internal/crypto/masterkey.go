package crypto

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
)

const masterKeySize = 32

// EnsureMasterKey creates secret.key with 32 random bytes if it does not exist.
func EnsureMasterKey(path string) (created bool, err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return false, fmt.Errorf("mkdir secret dir: %w", err)
	}
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}

	key := make([]byte, masterKeySize)
	if _, err := rand.Read(key); err != nil {
		return false, fmt.Errorf("rand: %w", err)
	}
	if err := os.WriteFile(path, key, 0o600); err != nil {
		return false, fmt.Errorf("write secret: %w", err)
	}
	return true, nil
}
