package securitypolicy

import "strings"

// RouteSnapshot is resolved edge settings for one proxied service host.
type RouteSnapshot struct {
	Headers       map[string]string
	CSP           string
	CSPReportOnly bool
	BasicUser     string
	BasicHash     string
	Allowlist     []string
	Maintenance   bool
}

// SnapshotForRoute builds proxy render input from instance + optional site row.
func SnapshotForRoute(inst Instance, site Site) RouteSnapshot {
	return RouteSnapshot{
		Headers:       ResolvedHeaders(inst, site),
		CSP:           strings.TrimSpace(site.CSP),
		CSPReportOnly: site.CSPReportOnly,
		BasicUser:     strings.TrimSpace(site.BasicUser),
		BasicHash:     strings.TrimSpace(site.BasicHash),
		Allowlist:     append([]string(nil), site.Allowlist...),
		Maintenance:   site.Maintenance,
	}
}
