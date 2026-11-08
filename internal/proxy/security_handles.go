package proxy

import (
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/securitypolicy"
)

// SecurityRoute carries resolved edge controls for one hostname route.
type SecurityRoute struct {
	Snapshot securitypolicy.RouteSnapshot
}

func policyHeadersHandler(hdrs map[string]string) map[string]any {
	set := make(map[string]any, len(hdrs))
	for k, v := range hdrs {
		if strings.TrimSpace(v) == "" {
			continue
		}
		set[k] = []string{v}
	}
	if len(set) == 0 {
		return baselineHeadersHandler()
	}
	return map[string]any{
		"handler": "headers",
		"response": map[string]any{
			"set":      set,
			"deferred": true,
			"delete":   []string{"Server"},
		},
	}
}

func cspHeadersHandler(snap securitypolicy.RouteSnapshot) map[string]any {
	csp := strings.TrimSpace(snap.CSP)
	if csp == "" {
		return nil
	}
	name := "Content-Security-Policy"
	if snap.CSPReportOnly {
		name = "Content-Security-Policy-Report-Only"
	}
	return map[string]any{
		"handler": "headers",
		"response": map[string]any{
			"set": map[string]any{
				name: []string{csp},
			},
			"deferred": true,
		},
	}
}

func maintenanceHandler() map[string]any {
	body := "<!DOCTYPE html><html><head><title>Maintenance</title></head><body><h1>Maintenance</h1><p>This site is temporarily unavailable.</p></body></html>"
	return map[string]any{
		"handler":     "static_response",
		"status_code":   503,
		"headers":       map[string]any{"Content-Type": []string{"text/html; charset=utf-8"}},
		"body":          body,
		"close":         true,
	}
}

func basicAuthHandler(user, hash string) map[string]any {
	if user == "" || hash == "" {
		return nil
	}
	return map[string]any{
		"handler": "authentication",
		"providers": map[string]any{
			"http_basic": map[string]any{
				"hash": map[string]any{
					user: hash,
				},
			},
		},
	}
}

func ipDenyHandler(ranges []string) map[string]any {
	if len(ranges) == 0 {
		return nil
	}
	return map[string]any{
		"handler": "subroute",
		"routes": []any{
			map[string]any{
				"match": []any{
					map[string]any{
						"not": []any{
							map[string]any{
								"client_ip": map[string]any{
									"ranges": ranges,
								},
							},
						},
					},
				},
				"handle": []any{
					map[string]any{
						"handler":     "static_response",
						"status_code": 403,
						"body":        "Forbidden",
						"close":       true,
					},
				},
			},
		},
	}
}

// prependSecurity builds headers → encode → access → upstream handles.
func prependSecurity(snap securitypolicy.RouteSnapshot, upstream []any) []any {
	out := make([]any, 0, len(upstream)+6)
	out = append(out, policyHeadersHandler(snap.Headers))
	if h := cspHeadersHandler(snap); h != nil {
		out = append(out, h)
	}
	out = append(out, encodeHandler())
	if snap.Maintenance {
		out = append(out, maintenanceHandler())
		return out
	}
	if h := basicAuthHandler(snap.BasicUser, snap.BasicHash); h != nil {
		out = append(out, h)
	}
	if h := ipDenyHandler(snap.Allowlist); h != nil {
		out = append(out, h)
	}
	return append(out, upstream...)
}
