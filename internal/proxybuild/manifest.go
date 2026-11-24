package proxybuild

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed manifest.json
var manifestJSON []byte

// Manifest describes published proxybuild images.
type Manifest struct {
	Repository       string      `json:"repository"`
	CaddyVersion     string      `json:"caddy_version"`
	Cosign           CosignPolicy `json:"cosign"`
	DefaultImage     string      `json:"default_image"`
}

// CosignPolicy is used when verifying image signatures.
type CosignPolicy struct {
	CertificateIdentityRegexp string `json:"certificate_identity_regexp"`
	CertificateOIDCIssuer     string `json:"certificate_oidc_issuer"`
}

// LoadManifest returns the embedded release manifest.
func LoadManifest() (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(manifestJSON, &m); err != nil {
		return Manifest{}, err
	}
	if strings.TrimSpace(m.Repository) == "" {
		return Manifest{}, fmt.Errorf("proxybuild manifest: repository required")
	}
	if strings.TrimSpace(m.DefaultImage) == "" {
		m.DefaultImage = "caddy:2-alpine"
	}
	return m, nil
}

// ImageRefForVersion returns the tagged image reference for a dockyrd release version.
func ImageRefForVersion(panelVersion string) (string, error) {
	m, err := LoadManifest()
	if err != nil {
		return "", err
	}
	v := normalizeVersion(panelVersion)
	if v == "" || v == "dev" {
		return m.DefaultImage, nil
	}
	return m.Repository + ":v" + v, nil
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	return v
}