package proxy

import (
	"strings"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/securitypolicy"
)

func TestRenderStrictSecuritySnapshot(t *testing.T) {
	snap := securitypolicy.RouteSnapshot{
		Headers: map[string]string{
			"Strict-Transport-Security": "max-age=31536000",
			"X-Content-Type-Options":    "nosniff",
			"Permissions-Policy":        "accelerometer=()",
		},
		CSP:           "default-src 'self'",
		CSPReportOnly: true,
	}
	raw, err := Render(RenderInput{
		ACMEEmail: "ops@example.com",
		AppRoutes: []AppRoute{{
			Hostname:     "app.example.com",
			UpstreamDial: "dy-app:3000",
			Security:     &snap,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, frag := range []string{
		"max-age=31536000",
		"Content-Security-Policy-Report-Only",
		"default-src 'self'",
		"Permissions-Policy",
		"reverse_proxy",
	} {
		if !strings.Contains(s, frag) {
			t.Fatalf("missing %q in render", frag)
		}
	}
}

func TestRenderAccessCombo(t *testing.T) {
	snap := securitypolicy.RouteSnapshot{
		Headers: map[string]string{
			"X-Content-Type-Options": "nosniff",
		},
		BasicUser: "ops",
		BasicHash: "$2a$10$abcdefghijklmnopqrstuv",
		Allowlist: []string{"203.0.113.0/24"},
	}
	raw, err := Render(RenderInput{
		ACMEEmail: "ops@example.com",
		AppRoutes: []AppRoute{{
			Hostname:     "locked.example.com",
			UpstreamDial: "dy-locked:8080",
			Security:     &snap,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, frag := range []string{`"handler": "authentication"`, "203.0.113.0/24", "client_ip", "403"} {
		if !strings.Contains(s, frag) {
			t.Fatalf("missing %q", frag)
		}
	}
}
