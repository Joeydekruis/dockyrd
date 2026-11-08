package proxy

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

const acmeProbePath = "/.well-known/acme-challenge/dockyrd-probe"

// ACMEChallengeBlocked reports whether plain HTTP for the hostname redirects to HTTPS
// before a certificate exists (breaks Let's Encrypt HTTP-01).
func ACMEChallengeBlocked(hostname string) bool {
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return false
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1"+acmeProbePath, nil)
	if err != nil {
		return false
	}
	req.Host = hostname
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		return false
	}
	loc := strings.ToLower(strings.TrimSpace(resp.Header.Get("Location")))
	return strings.HasPrefix(loc, "https://")
}

// TLSHealthy reports whether :443 serves a certificate matching hostname locally.
func TLSHealthy(hostname string) bool {
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return true
	}
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(&dialer, "tcp", net.JoinHostPort("127.0.0.1", "443"), &tls.Config{
		ServerName:         hostname,
		InsecureSkipVerify: true,
	})
	if err != nil {
		return false
	}
	state := conn.ConnectionState()
	_ = conn.Close()
	if len(state.PeerCertificates) == 0 {
		return false
	}
	return certMatchesHostname(hostname, state.PeerCertificates[0])
}

// PlainHTTPOn443 reports whether something on :443 responds like cleartext HTTP (broken edge bootstrap).
func PlainHTTPOn443() bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", "443"), 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write([]byte("GET / HTTP/1.0\r\nHost: probe.local\r\n\r\n")); err != nil {
		return false
	}
	buf := make([]byte, 8)
	n, err := io.ReadAtLeast(conn, buf, 5)
	if err != nil || n < 5 {
		return false
	}
	return bytes.HasPrefix(buf[:n], []byte("HTTP/"))
}

func certMatchesHostname(hostname string, leaf *x509.Certificate) bool {
	if err := leaf.VerifyHostname(hostname); err == nil {
		return true
	}
	h := strings.ToLower(strings.TrimSuffix(hostname, "."))
	for _, name := range leaf.DNSNames {
		if strings.EqualFold(name, h) {
			return true
		}
	}
	return strings.EqualFold(leaf.Subject.CommonName, h)
}
