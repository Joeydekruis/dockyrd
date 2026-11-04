package staticsite

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	if err := os.MkdirAll(paths.ServiceRoot, 0o755); err != nil {
		return "", err
	}
	if err := linkCurrentRelease(paths.CurrentLink, releasePath); err != nil {
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
	if err := linkCurrentRelease(paths.CurrentLink, releasePath); err != nil {
		return "", err
	}
	return releasePath, nil
}

// linkCurrentRelease atomically points current at releasePath using a relative symlink target.
func linkCurrentRelease(currentLink, releasePath string) error {
	rel, err := filepath.Rel(filepath.Dir(currentLink), releasePath)
	if err != nil {
		return err
	}
	tmpLink := currentLink + ".tmp"
	_ = os.Remove(tmpLink)
	if err := os.Symlink(rel, tmpLink); err != nil {
		return err
	}
	return os.Rename(tmpLink, currentLink)
}

// EnsureRelativeCurrentLink rewrites legacy absolute current→release symlinks (dy-proxy mount safe).
func EnsureRelativeCurrentLink(sitesRoot, serviceSlug string) error {
	paths := PathsForService(sitesRoot, serviceSlug)
	target, err := os.Readlink(paths.CurrentLink)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !filepath.IsAbs(target) {
		return nil
	}
	if _, err := os.Stat(target); err != nil {
		return err
	}
	return linkCurrentRelease(paths.CurrentLink, target)
}

// CaddyMountSitesRoot is the bind mount target inside dy-proxy (see docker.EnsureProxy).
const CaddyMountSitesRoot = "/srv/sites"

// CurrentRoot returns the host path to the current release symlink (publish, doctor).
func CurrentRoot(sitesRoot, serviceSlug string) string {
	return PathsForService(sitesRoot, serviceSlug).CurrentLink
}

// CaddyFileServerRoot is the file_server root path as seen inside dy-proxy (via current symlink).
func CaddyFileServerRoot(serviceSlug string) string {
	return filepath.Join(CaddyMountSitesRoot, serviceSlug, "current")
}

// HostReleaseToCaddyRoot maps a host release_path under sitesRoot to dy-proxy /srv/sites/…
func HostReleaseToCaddyRoot(sitesRoot, hostReleasePath string) (string, bool) {
	if strings.TrimSpace(hostReleasePath) == "" {
		return "", false
	}
	root := filepath.Clean(sitesRoot)
	release := filepath.Clean(hostReleasePath)
	rel, err := filepath.Rel(root, release)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.Join(CaddyMountSitesRoot, rel), true
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
