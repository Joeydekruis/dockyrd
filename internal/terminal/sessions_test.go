package terminal

import "testing"

func TestSessionLimitPerUser(t *testing.T) {
	r := NewSessionRegistry()
	if err := r.TryStart("u1", "a"); err != nil {
		t.Fatal(err)
	}
	if err := r.TryStart("u1", "b"); err != nil {
		t.Fatal(err)
	}
	if err := r.TryStart("u1", "c"); err != nil {
		t.Fatal(err)
	}
	if err := r.TryStart("u1", "d"); err != ErrSessionLimit {
		t.Fatalf("got %v", err)
	}
	r.End("a")
	if err := r.TryStart("u1", "d"); err != nil {
		t.Fatalf("after end: %v", err)
	}
}
