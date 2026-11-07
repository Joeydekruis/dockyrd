package proxy

import (
	"strings"
	"testing"
)

func TestIPv4ToSSLIPBase(t *testing.T) {
	got, err := IPv4ToSSLIPBase("203.0.113.5")
	if err != nil {
		t.Fatal(err)
	}
	if got != "203-0-113-5" {
		t.Fatalf("got %q", got)
	}
}

func TestPlatformHostnameShape(t *testing.T) {
	host, err := PlatformHostname("api", "203.0.113.5")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(host, ".203-0-113-5.sslip.io") {
		t.Fatalf("suffix mismatch: %q", host)
	}
	if !strings.HasPrefix(host, "api-") {
		t.Fatalf("prefix mismatch: %q", host)
	}
}

func TestPlatformHostnameForVerifiedWildcard(t *testing.T) {
	host, err := PlatformHostnameFor(PlatformHostInput{
		ServiceSlug:      "web-api",
		PublicIPv4:       "203.0.113.5",
		WildcardPattern:  "*.apps.clientx.nl",
		WildcardVerified: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(host, ".apps.clientx.nl") {
		t.Fatalf("suffix: %q", host)
	}
	if strings.Contains(host, "sslip.io") {
		t.Fatalf("should not use sslip: %q", host)
	}
}

func TestRenderContainsPanelRoute(t *testing.T) {
	raw, err := Render(RenderInput{
		ACMEEmail:     "ops@example.com",
		PanelHost:     "dockyrd.example.com",
		PanelUpstream: "127.0.0.1:8000",
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "dockyrd.example.com") || !strings.Contains(s, "127.0.0.1:8000") {
		t.Fatalf("missing panel route in %s", s)
	}
}

func TestRenderContainsRoute(t *testing.T) {
	raw, err := Render(RenderInput{
		ACMEEmail: "ops@example.com",
		AppRoutes: []AppRoute{{
			Hostname:     "api.clientx.nl",
			UpstreamDial: "dy-api-01jc9k:3000",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "api.clientx.nl") || !strings.Contains(s, "dy-api-01jc9k:3000") {
		t.Fatalf("missing route in %s", s)
	}
	if !strings.Contains(s, "reverse_proxy") {
		t.Fatal("expected reverse_proxy handler")
	}
}

func TestRenderTrustedProxiesCloudflare(t *testing.T) {
	raw, err := Render(RenderInput{
		ACMEEmail: "ops@example.com",
		Trusted: TrustedProxyConfig{
			Mode:           "cloudflare",
			Ranges:         []string{"173.245.48.0/20"},
			ClientIPHeader: "CF-Connecting-IP",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "trusted_proxies") || !strings.Contains(s, "CF-Connecting-IP") {
		t.Fatalf("missing trusted proxy config: %s", s)
	}
}

func TestRenderBaselineHandlers(t *testing.T) {
	raw, err := Render(RenderInput{
		ACMEEmail: "ops@example.com",
		AppRoutes: []AppRoute{{
			Hostname:     "app.example.com",
			UpstreamDial: "dy-app:3000",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, frag := range []string{`"handler": "encode"`, `Strict-Transport-Security`, `nosniff`, `read_header_timeout`, `disable_redirects`, `acme-challenge`} {
		if !strings.Contains(s, frag) {
			t.Fatalf("missing %q in render output", frag)
		}
	}
}

func TestRenderStaticFileServerRoot(t *testing.T) {
	raw, err := Render(RenderInput{
		ACMEEmail: "ops@example.com",
		StaticRoutes: []StaticRoute{{
			Hostname:   "marketing.example.com",
			Root:       "/srv/sites/marketing/current",
			IndexNames: []string{"index.html"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "/srv/sites/marketing/current") {
		t.Fatalf("expected container sites root in %s", s)
	}
	if !strings.Contains(s, "file_server") {
		t.Fatal("expected file_server handler")
	}
}
