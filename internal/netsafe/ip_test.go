package netsafe

import (
	"net"
	"testing"
)

func TestBlockedOutboundIP(t *testing.T) {
	t.Parallel()
	cases := []struct {
		ip   string
		want bool
	}{
		{"127.0.0.1", true},
		{"10.0.0.1", true},
		{"192.168.1.1", true},
		{"169.254.169.254", true},
		{"100.64.0.1", true},
		{"8.8.8.8", false},
	}
	for _, tc := range cases {
		ip := net.ParseIP(tc.ip)
		if got := BlockedOutboundIP(ip); got != tc.want {
			t.Fatalf("%s blocked=%v want %v", tc.ip, got, tc.want)
		}
	}
}
