package netutil

import "net"

// IsPublicIPv4 reports whether ip is a globally routable IPv4 address.
func IsPublicIPv4(ip net.IP) bool {
	ip = ip.To4()
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	if ip[0] == 10 {
		return false
	}
	if ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31 {
		return false
	}
	if ip[0] == 192 && ip[1] == 168 {
		return false
	}
	if ip[0] == 100 && ip[1] >= 64 && ip[1] <= 127 {
		return false
	}
	return true
}

// PrimaryIPv4 returns the best local IPv4 for sslip.io platform hostnames and DNS hints.
func PrimaryIPv4() (string, bool) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", false
	}
	var fallback []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil || ip4.IsLoopback() || ip4.IsLinkLocalUnicast() {
				continue
			}
			s := ip4.String()
			if IsPublicIPv4(ip4) {
				return s, true
			}
			fallback = append(fallback, s)
		}
	}
	if len(fallback) > 0 {
		return fallback[0], true
	}
	return "", false
}
