package setup

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

type bannerDisplayMode int

const (
	bannerLoopback bannerDisplayMode = iota
	bannerDirectURL
	bannerManual
)

func listenPort(listen string) string {
	if listen == "" {
		return "8000"
	}
	if !strings.Contains(listen, ":") {
		return listen
	}
	_, port, err := net.SplitHostPort(listen)
	if err != nil {
		if strings.HasPrefix(listen, ":") {
			return strings.TrimPrefix(listen, ":")
		}
		return "8000"
	}
	if port == "" {
		return "8000"
	}
	return port
}

func listenBindHost(listen string) string {
	if listen == "" || strings.HasPrefix(listen, ":") {
		return ""
	}
	if !strings.Contains(listen, ":") {
		return listen
	}
	host, _, err := net.SplitHostPort(listen)
	if err != nil {
		return ""
	}
	return host
}

func isPublicIPv4(ip net.IP) bool {
	ip = ip.To4()
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	// RFC1918 and CGNAT-ish ranges treated as non-public for banner preference.
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

func primaryIPv4() (string, bool) {
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
			if isPublicIPv4(ip4) {
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

func resolveBannerDisplay(listen string, detect func() (string, bool)) (bannerDisplayMode, string) {
	bind := listenBindHost(listen)
	if bind == "127.0.0.1" || bind == "localhost" {
		return bannerLoopback, "127.0.0.1"
	}
	if ip, ok := detect(); ok {
		return bannerDirectURL, ip
	}
	return bannerManual, ""
}

func writeSetupBanner(w io.Writer, token string, expiresAt time.Time, listen string, detect func() (string, bool)) {
	if token == "" {
		fmt.Fprintln(w, "Dockyard is initialized (setup token unchanged).")
		fmt.Fprintln(w, "  Lost it?  dockyard setup-token")
		return
	}
	fmt.Fprintln(w, "✓ Dockyard initialized.")
	port := listenPort(listen)
	mode, host := resolveBannerDisplay(listen, detect)
	fmt.Fprintln(w)
	switch mode {
	case bannerLoopback:
		fmt.Fprintf(w, "  Setup token:  %s\n", token)
		fmt.Fprintf(w, "  Open:  http://127.0.0.1:%s/setup?token=%s\n", port, token)
	case bannerDirectURL:
		fmt.Fprintf(w, "  Setup URL:  http://%s:%s/setup?token=%s\n", host, port, token)
	case bannerManual:
		fmt.Fprintf(w, "  Setup token:  %s\n", token)
		fmt.Fprintf(w, "  Open in your browser:  http://<this-server-ip>:%s/setup?token=%s\n", port, token)
		fmt.Fprintln(w, "  Use this machine's public IP (not 127.0.0.1 from your laptop).")
	}
	fmt.Fprintf(w, "  Token valid until %s UTC (single use).\n", expiresAt.Format(time.RFC3339))
	fmt.Fprintln(w, "  Lost it?  dockyard setup-token")
}

// PrintInitBanner writes the post-install message to stdout.
func PrintInitBanner(token string, expiresAt time.Time, listen string) {
	writeSetupBanner(os.Stdout, token, expiresAt, listen, primaryIPv4)
}
