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

func TestRenderContainsPanelRoute(t *testing.T) {
	raw, err := Render(RenderInput{
		ACMEEmail:     "ops@example.com",
		PanelHost:     "dockyrd.example.com",
		PanelUpstream: "host.docker.internal:8000",
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "dockyrd.example.com") || !strings.Contains(s, "host.docker.internal:8000") {
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
