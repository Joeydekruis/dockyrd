package staticsite

import "testing"

func TestCaddyFileServerRoot(t *testing.T) {
	got := CaddyFileServerRoot("dockyrd")
	want := "/srv/sites/dockyrd/current"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
