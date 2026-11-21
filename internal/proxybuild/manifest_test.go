package proxybuild

import "testing"

func TestImageRefForVersion(t *testing.T) {
	ref, err := ImageRefForVersion("0.7.21")
	if err != nil {
		t.Fatal(err)
	}
	want := "ghcr.io/joeydekruis/dockyrd-proxy:v0.7.21"
	if ref != want {
		t.Fatalf("got %q want %q", ref, want)
	}
	ref, err = ImageRefForVersion("dev")
	if err != nil {
		t.Fatal(err)
	}
	if ref != "caddy:2-alpine" {
		t.Fatalf("dev default: got %q", ref)
	}
}