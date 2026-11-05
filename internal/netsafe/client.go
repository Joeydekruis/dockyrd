package netsafe

import (
	"context"
	"net"
	"net/http"
	"time"
)

// ClientOptions configures outbound HTTP from the control plane.
type ClientOptions struct {
	// AllowPrivateTargets permits RFC1918/ULA when true (owner opt-in); link-local/metadata stays blocked.
	AllowPrivateTargets bool
	Timeout             time.Duration
}

// HTTPClient returns an http.Client that refuses redirects and blocks SSRF targets at dial time.
func HTTPClient(opts ClientOptions) *http.Client {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	transport := &http.Transport{
		ForceAttemptHTTP2: true,
		MaxIdleConns:      8,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}
			if len(ips) == 0 {
				return nil, &net.OpError{Op: "dial", Net: network, Err: errBlockedIP{}}
			}
			for _, ipa := range ips {
				if !dialAllowed(ipa.IP, opts.AllowPrivateTargets) {
					return nil, &net.OpError{Op: "dial", Net: network, Err: errBlockedIP{}}
				}
			}
			target := ips[0].IP.String()
			return dialer.DialContext(ctx, network, net.JoinHostPort(target, port))
		},
	}
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport:     transport,
	}
}

func dialAllowed(ip net.IP, allowPrivate bool) bool {
	if ip == nil {
		return false
	}
	if alwaysBlockedIP(ip) {
		return false
	}
	if BlockedOutboundIP(ip) && !allowPrivate {
		return false
	}
	return true
}

func alwaysBlockedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil && ip4[0] == 169 && ip4[1] == 254 {
		return true
	}
	return false
}

type errBlockedIP struct{}

func (errBlockedIP) Error() string { return "blocked outbound IP" }
