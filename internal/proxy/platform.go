package proxy

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"regexp"
	"strings"
)

var slugPart = regexp.MustCompile(`[^a-z0-9-]+`)

// IPv4ToSSLIPBase turns 203.0.113.5 into 203-0-113-5 for sslip.io hostnames.
func IPv4ToSSLIPBase(ip string) (string, error) {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return "", fmt.Errorf("invalid ipv4: %q", ip)
	}
	v4 := parsed.To4()
	if v4 == nil {
		return "", fmt.Errorf("not an ipv4: %q", ip)
	}
	return fmt.Sprintf("%d-%d-%d-%d", v4[0], v4[1], v4[2], v4[3]), nil
}

// PlatformHostname builds <slug>-<rand4>.<ip-dashes>.sslip.io .
func PlatformHostname(serviceSlug, publicIPv4 string) (string, error) {
	base, err := IPv4ToSSLIPBase(publicIPv4)
	if err != nil {
		return "", err
	}
	slug := sanitizeSlugPart(serviceSlug)
	rand4, err := randomHex4()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s.%s.sslip.io", slug, rand4, base), nil
}

func sanitizeSlugPart(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugPart.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "app"
	}
	if len(s) > 32 {
		return s[:32]
	}
	return s
}

func randomHex4() (string, error) {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
