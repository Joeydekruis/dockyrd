package staticsite

import "testing"

func TestCaddyFileServerRoot(t *testing.T) {
	got := CaddyFileServerRoot("dockyard")
	want := "/srv/sites/dockyard/current"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
