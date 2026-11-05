package dns

import (
	"context"
	"net"
	"strings"
)

// EvaluateARecord compares the hostname's A records to expectedIPv4.
// Returns active, pending, or wrong_target (aligned with domains.dns_status).
func EvaluateARecord(ctx context.Context, hostname, expectedIPv4 string) string {
	hostname = strings.TrimSpace(strings.ToLower(hostname))
	expectedIPv4 = strings.TrimSpace(expectedIPv4)
	if hostname == "" {
		return "pending"
	}
	if expectedIPv4 == "" || expectedIPv4 == "127.0.0.1" {
		return "pending"
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip4", hostname)
	if err != nil || len(ips) == 0 {
		return "pending"
	}
	for _, ip := range ips {
		if ip.To4() != nil && ip.String() == expectedIPv4 {
			return "active"
		}
	}
	return "wrong_target"
}
