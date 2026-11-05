package netutil

import (
	"net"
	"testing"
)

func TestIsPublicIPv4(t *testing.T) {
	pub := net.ParseIP("203.0.113.5")
	priv := net.ParseIP("10.0.0.1")
	if !IsPublicIPv4(pub) {
		t.Fatal("expected public")
	}
	if IsPublicIPv4(priv) {
		t.Fatal("expected private")
	}
}
