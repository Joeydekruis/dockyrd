package build

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// DefaultRailpackVersion pins the CLI release and the BuildKit frontend tag.
	DefaultRailpackVersion = "0.2.2"
	railpackRepo           = "railwayapp/railpack"
)

// DefaultRailpackFrontend is the BuildKit syntax image for railpack plans.
const DefaultRailpackFrontend = "ghcr.io/railwayapp/railpack-frontend:v" + DefaultRailpackVersion

// EnsureRailpackCLI installs the railpack binary when it is not on PATH.
func EnsureRailpackCLI(ctx context.Context) error {
	if os.Getenv("dockyrd_SKIP_RAILPACK") == "1" {
		return nil
	}
	if _, err := exec.LookPath("railpack"); err == nil {
		return nil
	}
	dest, err := railpackInstallDest()
	if err != nil {
		return err
	}
	target, err := railpackAssetTarget()
	if err != nil {
		return err
	}
	version := strings.TrimSpace(os.Getenv("dockyrd_RAILPACK_VERSION"))
	if version == "" {
		version = DefaultRailpackVersion
	}
	url := fmt.Sprintf(
		"https://github.com/%s/releases/download/v%s/railpack-v%s-%s.tar.gz",
		railpackRepo, version, version, target,
	)
	tmp, err := os.MkdirTemp("", "dockyrd-railpack-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	archive := filepath.Join(tmp, "railpack.tar.gz")
	if err := downloadRailpack(ctx, url, archive); err != nil {
		return fmt.Errorf("download railpack v%s: %w", version, err)
	}
	if err := extractRailpackBinary(archive, dest); err != nil {
		return err
	}
	if _, err := exec.LookPath("railpack"); err != nil {
		if _, statErr := os.Stat(dest); statErr == nil {
			return nil
		}
		return fmt.Errorf("railpack installed to %s but not found on PATH", dest)
	}
	return nil
}

func railpackInstallDest() (string, error) {
	if p := strings.TrimSpace(os.Getenv("dockyrd_RAILPACK_BIN")); p != "" {
		return p, nil
	}
	dir := "/usr/local/bin"
	if d := strings.TrimSpace(os.Getenv("RAILPACK_BIN_DIR")); d != "" {
		dir = d
	}
	dest := filepath.Join(dir, "railpack")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", err
	}
	if fi, err := os.Stat(dest); err == nil && !fi.IsDir() {
		return dest, nil
	}
	if err := checkWritableDir(filepath.Dir(dest)); err != nil {
		return "", fmt.Errorf("cannot install railpack: %w (set RAILPACK_BIN_DIR or run as root)", err)
	}
	return dest, nil
}

func checkWritableDir(dir string) error {
	f, err := os.CreateTemp(dir, ".dockyrd-write-*")
	if err != nil {
		return err
	}
	name := f.Name()
	_ = f.Close()
	return os.Remove(name)
}

func railpackAssetTarget() (string, error) {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "arm64"
	default:
		return "", fmt.Errorf("unsupported GOARCH for railpack: %s", runtime.GOARCH)
	}
	var platform string
	switch runtime.GOOS {
	case "linux":
		platform = "unknown-linux-musl"
	case "darwin":
		platform = "apple-darwin"
	default:
		return "", fmt.Errorf("unsupported GOOS for railpack: %s", runtime.GOOS)
	}
	return arch + "-" + platform, nil
}

func downloadRailpack(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "dockyrd-railpack-install")
	client := &http.Client{Timeout: 5 * time.Minute}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %s", res.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, io.LimitReader(res.Body, 200<<20))
	return err
}

func extractRailpackBinary(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		base := filepath.Base(hdr.Name)
		if base != "railpack" {
			continue
		}
		tmp := dest + ".new"
		out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			_ = out.Close()
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
		if err := os.Rename(tmp, dest); err != nil {
			return err
		}
		return os.Chmod(dest, 0o755)
	}
	return fmt.Errorf("railpack binary not found in archive")
}
