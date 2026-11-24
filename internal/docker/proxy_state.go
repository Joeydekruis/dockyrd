package docker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

const proxyImageStateFile = "proxy-image.json"

// ProxyImageState persists the active dy-proxy image reference on the host.
type ProxyImageState struct {
	Image     string `json:"image"`
	Digest    string `json:"digest,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
	Proxybuild bool  `json:"proxybuild,omitempty"`
}

// EffectiveProxyImage returns the image ref for dy-proxy (state file or default stock Caddy).
func EffectiveProxyImage(cfg config.Config) string {
	st, err := ReadProxyImageState(cfg)
	if err != nil || strings.TrimSpace(st.Image) == "" {
		return ProxyImage
	}
	return strings.TrimSpace(st.Image)
}

// ReadProxyImageState loads proxy-image.json from the caddy data dir.
func ReadProxyImageState(cfg config.Config) (ProxyImageState, error) {
	path := proxyImageStatePath(cfg)
	data, err := os.ReadFile(path)
	if err != nil {
		return ProxyImageState{}, err
	}
	var st ProxyImageState
	if err := json.Unmarshal(data, &st); err != nil {
		return ProxyImageState{}, err
	}
	return st, nil
}

// WriteProxyImageState saves the active proxy image ref.
func WriteProxyImageState(cfg config.Config, st ProxyImageState) error {
	st.Image = strings.TrimSpace(st.Image)
	if st.Image == "" {
		return os.Remove(proxyImageStatePath(cfg))
	}
	if st.UpdatedAt == 0 {
		st.UpdatedAt = time.Now().Unix()
	}
	if st.Image != ProxyImage && !st.Proxybuild {
		st.Proxybuild = strings.Contains(st.Image, "dockyrd-proxy")
	}
	path := proxyImageStatePath(cfg)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func proxyImageStatePath(cfg config.Config) string {
	return filepath.Join(cfg.CaddyDir(), proxyImageStateFile)
}