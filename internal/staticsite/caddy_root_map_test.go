package staticsite

import "testing"

func TestHostReleaseToCaddyRoot(t *testing.T) {
	sitesRoot := "/var/lib/dockyard/sites"
	release := "/var/lib/dockyard/sites/dockyard/releases/01ABC"
	got, ok := HostReleaseToCaddyRoot(sitesRoot, release)
	if !ok {
		t.Fatal("expected ok")
	}
	want := "/srv/sites/dockyard/releases/01ABC"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestHostReleaseToCaddyRoot_rejectsOutsideRoot(t *testing.T) {
	_, ok := HostReleaseToCaddyRoot("/var/lib/dockyard/sites", "/etc/passwd")
	if ok {
		t.Fatal("expected false")
	}
}
