package proxybuild

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// VerifyImageSignature checks the image with cosign when available and policy is set.
func VerifyImageSignature(ctx context.Context, imageRef string) error {
	imageRef = strings.TrimSpace(imageRef)
	if imageRef == "" {
		return errors.New("image ref required")
	}
	if strings.HasPrefix(imageRef, "caddy:") {
		// Stock Caddy — no Dockyrd cosign identity.
		return nil
	}
	if _, err := exec.LookPath("cosign"); err != nil {
		return nil
	}
	m, err := LoadManifest()
	if err != nil {
		return err
	}
	id := strings.TrimSpace(m.Cosign.CertificateIdentityRegexp)
	iss := strings.TrimSpace(m.Cosign.CertificateOIDCIssuer)
	if id == "" || iss == "" {
		return nil
	}
	cmd := exec.CommandContext(ctx, "cosign", "verify", imageRef,
		"--certificate-identity-regexp", id,
		"--certificate-oidc-issuer", iss,
	)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cosign verify %s: %w (%s)", imageRef, err, strings.TrimSpace(string(out)))
	}
	return nil
}