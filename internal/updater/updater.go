package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
)

// Options tune update behavior.
type Options struct {
	Repo          string
	TargetVersion string // empty = latest from GitHub
	DryRun        bool
	BinaryPath    string
	Quiet         bool // suppress progress lines (API apply)
}

// ApplyResult is returned after a successful install.
type ApplyResult struct {
	InstalledTag string
}

// CheckResult is returned by Check.
type CheckResult struct {
	Current string
	Latest  Release
	Newer   bool
}

// Check queries GitHub for a newer release.
func Check(ctx context.Context, current string, opts Options) (CheckResult, error) {
	rel, newer, err := FetchLatest(ctx, opts.Repo, current)
	if err != nil {
		return CheckResult{}, err
	}
	return CheckResult{Current: current, Latest: rel, Newer: newer}, nil
}

// Apply downloads and installs a release (Linux only; see apply_linux.go).
func Apply(ctx context.Context, cfg config.Config, current string, opts Options) (ApplyResult, error) {
	snap := DefaultReleaseCache.Snapshot()
	var check CheckResult
	var err error
	if snap.HasData && snap.Err == nil {
		check = snap.Result
	} else {
		check, err = Check(ctx, current, opts)
		if err != nil {
			return ApplyResult{}, err
		}
	}
	rel := check.Latest
	if opts.TargetVersion != "" {
		if CompareVersions(current, opts.TargetVersion) >= 0 {
			return ApplyResult{}, fmt.Errorf("already at or past %s", opts.TargetVersion)
		}
		if CompareVersions(rel.Tag, opts.TargetVersion) != 0 && normalizeTag(opts.TargetVersion) != rel.Tag {
			return ApplyResult{}, fmt.Errorf("latest release is %s, not %s (pin not implemented)", rel.Tag, opts.TargetVersion)
		}
	} else if !check.Newer {
		return ApplyResult{}, fmt.Errorf("already up to date (%s)", current)
	}

	if opts.DryRun || os.Getenv("DOCKYARD_UPDATE_DRY_RUN") == "1" {
		if !opts.Quiet {
			fmt.Printf("dry-run: would install %s from %s\n", rel.Tag, rel.AssetName)
		}
		return ApplyResult{InstalledTag: rel.Tag}, nil
	}

	tmpDir, err := os.MkdirTemp("", "dockyard-update-*")
	if err != nil {
		return ApplyResult{}, err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, rel.AssetName)
	if err := downloadFile(ctx, rel.AssetURL, archivePath); err != nil {
		return ApplyResult{}, err
	}
	if rel.ChecksumURL != "" {
		if err := verifyChecksums(ctx, rel.ChecksumURL, rel.AssetName, archivePath); err != nil {
			return ApplyResult{}, err
		}
	}
	binPath, err := extractDockyardBinary(archivePath, tmpDir)
	if err != nil {
		return ApplyResult{}, err
	}

	backupPath := db.PreUpdateBackupPath(cfg.DataDir, time.Now())
	if _, err := os.Stat(cfg.DatabasePath()); err == nil {
		if err := db.VacuumIntoBackup(cfg.DatabasePath(), backupPath); err != nil {
			return ApplyResult{}, fmt.Errorf("pre-update backup: %w", err)
		}
		if !opts.Quiet {
			fmt.Fprintf(os.Stderr, "Database backed up to %s\n", backupPath)
		}
	}

	dest := opts.BinaryPath
	if dest == "" {
		dest = os.Getenv("DOCKYARD_BIN")
	}
	if dest == "" {
		dest = "/usr/local/bin/dockyard"
	}
	if err := installBinary(binPath, dest); err != nil {
		return ApplyResult{}, err
	}
	if !opts.Quiet {
		fmt.Printf("Installed %s to %s\n", rel.Tag, dest)
		fmt.Println("Restart the dockyard service: systemctl restart dockyard")
	}
	return ApplyResult{InstalledTag: rel.Tag}, nil
}

func downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 10 * time.Minute}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", url, res.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, res.Body)
	return err
}

func verifyChecksums(ctx context.Context, checksumsURL, assetName, archivePath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumsURL, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	want := parseChecksumLine(string(body), assetName)
	if want == "" {
		return nil
	}
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	got := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(got, want) {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	return nil
}

func parseChecksumLine(body, assetName string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[len(parts)-1] == assetName {
			return parts[0]
		}
	}
	return ""
}

func extractDockyardBinary(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		base := filepath.Base(hdr.Name)
		if base != "dockyard" && !strings.HasPrefix(base, "dockyard") {
			continue
		}
		out := filepath.Join(destDir, "dockyard")
		w, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(w, tr); err != nil {
			_ = w.Close()
			return "", err
		}
		if err := w.Close(); err != nil {
			return "", err
		}
		return out, nil
	}
	return "", fmt.Errorf("dockyard binary not found in archive")
}
