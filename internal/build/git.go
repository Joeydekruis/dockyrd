package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneShallow checks out one branch into dest (must not exist).
func CloneShallow(ctx context.Context, url, branch, dest string) error {
	url = strings.TrimSpace(url)
	branch = strings.TrimSpace(branch)
	if url == "" {
		return fmt.Errorf("git url required")
	}
	if branch == "" {
		branch = "main"
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
		return err
	}
	args := []string{
		"clone",
		"--depth", "1",
		"--branch", branch,
		"--single-branch",
		url,
		dest,
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
