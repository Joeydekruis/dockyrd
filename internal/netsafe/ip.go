package netsafe

import (
	"net"
)

// BlockedOutboundIP reports whether ip must not be dialed for user-configured webhooks.
func BlockedOutboundIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	ip = ip.To16()
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		return blockedIPv4(ip4)
	}
	return blockedIPv6(ip)
}

func blockedIPv4(ip net.IP) bool {
	// 0.0.0.0/8, 10/8, 127/8, 169.254/16, 172.16/12, 192.168/16, 100.64/10 (CGNAT)
	if ip[0] == 0 || ip[0] == 10 || ip[0] == 127 {
		return true
	}
	if ip[0] == 169 && ip[1] == 254 {
		return true
	}
	if ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31 {
		return true
	}
	if ip[0] == 192 && ip[1] == 168 {
		return true
	}
	if ip[0] == 100 && ip[1] >= 64 && ip[1] <= 127 {
		return true
	}
	return false
}

func blockedIPv6(ip net.IP) bool {
	// ULA fc00::/7, link-local fe80::/10 handled by IsLinkLocalUnicast
	if ip[0]&0xfe == 0xfc {
		return true
	}
	if ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
		return true
	}
	return false
}
