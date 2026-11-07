package setup

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/netutil"
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
		fmt.Fprintln(w, "dockyrd is initialized (setup token unchanged).")
		fmt.Fprintln(w, "  Lost it?  dockyrd setup-token")
		return
	}
	fmt.Fprintln(w, "✓ dockyrd initialized.")
	port := listenPort(listen)
	mode, host := resolveBannerDisplay(listen, detect)
	panelHost := instance.PanelHostname(installerPublicURL())
	fmt.Fprintln(w)
	if panelHost != "" {
		fmt.Fprintf(w, "  Setup URL:  https://%s/setup?token=%s\n", panelHost, token)
		fmt.Fprintln(w, "  (DNS must already point at this server; ACME needs ports 80 and 443 open.)")
	}
	switch mode {
	case bannerLoopback:
		fmt.Fprintf(w, "  Setup token:  %s\n", token)
		fmt.Fprintf(w, "  Open:  http://127.0.0.1:%s/setup?token=%s\n", port, token)
	case bannerDirectURL:
		if panelHost == "" {
			fmt.Fprintf(w, "  Setup URL:  http://%s:%s/setup?token=%s\n", host, port, token)
		}
	case bannerManual:
		fmt.Fprintf(w, "  Setup token:  %s\n", token)
		fmt.Fprintf(w, "  Open in your browser:  http://<this-server-ip>:%s/setup?token=%s\n", port, token)
		fmt.Fprintln(w, "  Use this machine's public IP (not 127.0.0.1 from your laptop).")
	}
	fmt.Fprintf(w, "  Token valid until %s UTC (single use).\n", expiresAt.Format(time.RFC3339))
	fmt.Fprintln(w, "  Lost it?  dockyrd setup-token")
}

func installerPublicURL() string {
	for _, key := range []string{"dockyrd_PUBLIC_URL", "DOCKYARD_PUBLIC_URL"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

// PrintInitBanner writes the post-install message to stdout.
func PrintInitBanner(token string, expiresAt time.Time, listen string) {
	writeSetupBanner(os.Stdout, token, expiresAt, listen, netutil.PrimaryIPv4)
}
