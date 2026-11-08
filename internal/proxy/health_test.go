package proxy

import "testing"

func TestACMEChallengeBlocked_emptyHost(t *testing.T) {
	if ACMEChallengeBlocked("") {
		t.Fatal("expected false for empty host")
	}
}

func TestTLSHealthy_emptyHost(t *testing.T) {
	if !TLSHealthy("") {
		t.Fatal("expected true for empty host (skip check)")
	}
}
