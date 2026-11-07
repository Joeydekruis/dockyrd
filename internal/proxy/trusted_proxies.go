package proxy

import (
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/securitypolicy"
)

// TrustedProxyConfig is server-level client IP trust (briefing §4.5 / P1d).
type TrustedProxyConfig struct {
	Mode          string
	Ranges        []string
	ClientIPHeader string
}

// TrustedFromInstance maps stored policy to proxy render input.
func TrustedFromInstance(inst securitypolicy.Instance) TrustedProxyConfig {
	mode := strings.ToLower(strings.TrimSpace(inst.CDNMode))
	if mode == "" || mode == "direct" {
		return TrustedProxyConfig{Mode: "direct"}
	}
	ranges := append([]string(nil), inst.CDNRanges...)
	header := strings.TrimSpace(inst.CDNHeader)
	if mode == "cloudflare" && header == "" {
		header = "CF-Connecting-IP"
	}
	return TrustedProxyConfig{
		Mode:           mode,
		Ranges:         ranges,
		ClientIPHeader: header,
	}
}

func applyTrustedProxies(server map[string]any, cfg TrustedProxyConfig) {
	if cfg.Mode == "" || cfg.Mode == "direct" || len(cfg.Ranges) == 0 {
		return
	}
	tp := map[string]any{
		"source": "static",
		"ranges": cfg.Ranges,
	}
	server["trusted_proxies"] = tp
	if cfg.ClientIPHeader != "" {
		server["client_ip_headers"] = []string{cfg.ClientIPHeader}
	}
	server["trusted_proxies_strict"] = true
}
