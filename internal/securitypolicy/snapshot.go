package securitypolicy

import "strings"

// RouteSnapshot is resolved edge settings for one proxied service host.
type RouteSnapshot struct {
	Headers            map[string]string
	CSP                string
	CSPReportOnly      bool
	BasicUser          string
	BasicHash          string
	Allowlist          []string
	Maintenance        bool
	WAFMode            WAFMode // effective mode (off|detect|block)
	WAFPathExceptions  []string
	ProxyWAFInstalled  bool
	AllowWAF           bool // false for static sites — never emit coraza_waf
}

// SnapshotForRoute builds proxy render input from instance + optional site row.
// Callers must set ProxyWAFInstalled / AllowWAF for edge emit gates.
func SnapshotForRoute(inst Instance, site Site) RouteSnapshot {
	return RouteSnapshot{
		Headers:           ResolvedHeaders(inst, site),
		CSP:               strings.TrimSpace(site.CSP),
		CSPReportOnly:     site.CSPReportOnly,
		BasicUser:         strings.TrimSpace(site.BasicUser),
		BasicHash:         strings.TrimSpace(site.BasicHash),
		Allowlist:         append([]string(nil), site.Allowlist...),
		Maintenance:       site.Maintenance,
		WAFMode:           EffectiveWAFMode(inst, site),
		WAFPathExceptions: append([]string(nil), site.WAFPathExceptions...),
		AllowWAF:          true,
	}
}
