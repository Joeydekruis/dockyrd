package staticsite

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ReleasePaths describes on-disk layout for one static site service.
type ReleasePaths struct {
	ServiceRoot string
	ReleasesDir string
	CurrentLink string
}

// PathsForService returns directory layout under sitesRoot.
func PathsForService(sitesRoot, serviceSlug string) ReleasePaths {
	root := filepath.Join(sitesRoot, serviceSlug)
	return ReleasePaths{
		ServiceRoot: root,
		ReleasesDir: filepath.Join(root, "releases"),
		CurrentLink: filepath.Join(root, "current"),
	}
}

// PublishRelease copies built output into releases/<deploymentID> and atomically updates current.
func PublishRelease(outputDir, sitesRoot, serviceSlug, deploymentID string) (releasePath string, err error) {
	paths := PathsForService(sitesRoot, serviceSlug)
	releasePath = filepath.Join(paths.ReleasesDir, deploymentID)
	if err := os.RemoveAll(releasePath); err != nil {
		return "", err
	}
	if err := os.MkdirAll(releasePath, 0o755); err != nil {
		return "", err
	}
	if err := copyTree(outputDir, releasePath); err != nil {
		return "", err
	}
	if err := os.MkdirAll(paths.ServiceRoot, 0o750); err != nil {
		return "", err
	}
	tmpLink := paths.CurrentLink + ".tmp"
	_ = os.Remove(tmpLink)
	if err := os.Symlink(releasePath, tmpLink); err != nil {
		return "", err
	}
	if err := os.Rename(tmpLink, paths.CurrentLink); err != nil {
		return "", err
	}
	return releasePath, nil
}

// ActivateRelease points current at an existing release directory (rollback).
func ActivateRelease(sitesRoot, serviceSlug, deploymentID string) (releasePath string, err error) {
	paths := PathsForService(sitesRoot, serviceSlug)
	releasePath = filepath.Join(paths.ReleasesDir, deploymentID)
	if _, err := os.Stat(releasePath); err != nil {
		return "", fmt.Errorf("release not found: %w", err)
	}
	tmpLink := paths.CurrentLink + ".tmp"
	_ = os.Remove(tmpLink)
	if err := os.Symlink(releasePath, tmpLink); err != nil {
		return "", err
	}
	if err := os.Rename(tmpLink, paths.CurrentLink); err != nil {
		return "", err
	}
	return releasePath, nil
}

// CurrentRoot returns the path Caddy should serve (symlink current).
func CurrentRoot(sitesRoot, serviceSlug string) string {
	return PathsForService(sitesRoot, serviceSlug).CurrentLink
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
