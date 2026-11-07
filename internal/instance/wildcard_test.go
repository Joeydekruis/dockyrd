package instance

import "testing"

func TestSuggestAppsWildcard(t *testing.T) {
	got := SuggestAppsWildcard("https://panel.clientx.nl")
	want := "*.apps.clientx.nl"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestWildcardSuffix(t *testing.T) {
	if WildcardSuffix("*.apps.example.com") != "apps.example.com" {
		t.Fatal("suffix strip")
	}
}

func TestProbeHostnameForWildcard(t *testing.T) {
	got := ProbeHostnameForWildcard("*.apps.clientx.nl")
	if got != "dockyrd-check.apps.clientx.nl" {
		t.Fatalf("got %q", got)
	}
}

func TestWildcardIsVerified(t *testing.T) {
	if WildcardIsVerified(Panel{Wildcard: "*.apps.x.com", WildcardVerifiedAt: 1}) != true {
		t.Fatal("expected verified")
	}
	if WildcardIsVerified(Panel{Wildcard: "*.apps.x.com"}) {
		t.Fatal("expected unverified")
	}
}
