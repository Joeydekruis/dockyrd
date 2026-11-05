package backup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/Joeydekruis/dockyrd/internal/docker"
)

const maxLogTailBytes = 32 << 10

// DigestFile returns the hex-encoded SHA-256 of path.
func DigestFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// VerifyPostgresDump runs pg_restore --list inside the database container.
func VerifyPostgresDump(ctx context.Context, dc *docker.Client, containerID, dumpPath string) error {
	if dc == nil || containerID == "" {
		return fmt.Errorf("docker unavailable for backup verification")
	}
	st, err := os.Stat(dumpPath)
	if err != nil {
		return err
	}
	const dest = "dy-verify.dump"
	f, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := dc.CopyFileToContainer(ctx, containerID, "/tmp", dest, f, st.Size()); err != nil {
		return fmt.Errorf("copy dump for verify: %w", err)
	}
	cmd := []string{"pg_restore", "--list", "/tmp/" + dest}
	if err := dc.ExecStream(ctx, containerID, cmd, io.Discard); err != nil {
		return fmt.Errorf("pg_restore --list: %w", err)
	}
	return nil
}

// TailBytes returns the last n bytes of s, capped for storage.
func TailBytes(s string, max int) string {
	if max <= 0 {
		max = maxLogTailBytes
	}
	if len(s) <= max {
		return s
	}
	return s[len(s)-max:]
}
