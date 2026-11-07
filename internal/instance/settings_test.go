package instance

import (
	"os"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

func TestPanelHostname(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://dockyrd.derisory.nl", "dockyrd.derisory.nl"},
		{"https://dockyrd.derisory.nl/", "dockyrd.derisory.nl"},
		{"http://panel.example.com:8000", "panel.example.com"},
		{"panel.example.com", "panel.example.com"},
		{"", ""},
		{"https://127.0.0.1", ""},
	}
	for _, tc := range tests {
		if got := PanelHostname(tc.in); got != tc.want {
			t.Errorf("PanelHostname(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestEffectivePublicURL(t *testing.T) {
	t.Setenv("dockyrd_PUBLIC_URL", "https://panel.env.test")
	cfg := config.Load()
	if got := EffectivePublicURL(cfg, Panel{PublicURL: "https://db.example.com"}); got != "https://db.example.com" {
		t.Fatalf("db wins: %q", got)
	}
	if got := EffectivePublicURL(cfg, Panel{}); got != "https://panel.env.test" {
		t.Fatalf("env fallback: %q", got)
	}
	_ = os.Unsetenv("dockyrd_PUBLIC_URL")
}

func TestEffectiveAcmeEmail(t *testing.T) {
	t.Setenv("dockyrd_ACME_EMAIL", "env@test.com")
	cfg := config.Load()
	if got := EffectiveAcmeEmail(cfg, Panel{AcmeEmail: "db@test.com"}); got != "db@test.com" {
		t.Fatalf("db wins: %q", got)
	}
	if got := EffectiveAcmeEmail(cfg, Panel{}); got != "env@test.com" {
		t.Fatalf("env fallback: %q", got)
	}
}
