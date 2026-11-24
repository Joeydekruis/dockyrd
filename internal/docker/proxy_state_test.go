package docker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

func TestProxyImageStateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir}
	if err := WriteProxyImageState(cfg, ProxyImageState{Image: "ghcr.io/joeydekruis/dockyrd-proxy:v0.7.21"}); err != nil {
		t.Fatal(err)
	}
	st, err := ReadProxyImageState(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if st.Image != "ghcr.io/joeydekruis/dockyrd-proxy:v0.7.21" || !st.Proxybuild {
		t.Fatalf("state: %+v", st)
	}
	if got := EffectiveProxyImage(cfg); got != st.Image {
		t.Fatalf("effective %q", got)
	}
	path := filepath.Join(cfg.CaddyDir(), proxyImageStateFile)
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}