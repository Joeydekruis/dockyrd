package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractReleaseBinaries(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	archive := filepath.Join(dir, "dockyard_0.0.0_linux_amd64.tar.gz")
	if err := writeTestArchive(archive, map[string]string{
		"dockyard": "#!/bin/sh\necho panel\n",
		"dy":       "#!/bin/sh\necho cli\n",
	}); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(dir, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dockyard, dy, err := extractReleaseBinaries(archive, outDir)
	if err != nil {
		t.Fatal(err)
	}
	if dockyard == "" || dy == "" {
		t.Fatalf("expected both binaries, got dockyard=%q dy=%q", dockyard, dy)
	}
	if _, err := os.Stat(dy); err != nil {
		t.Fatal(err)
	}
}

func writeTestArchive(path string, files map[string]string) error {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(body))}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
