package securitypolicy

// DefaultBaselineHeaders matches internal/proxy/baseline.go seed values.
func DefaultBaselineHeaders() map[string]HeaderEntry {
	return map[string]HeaderEntry{
		"Strict-Transport-Security": {Enabled: true, Value: "max-age=15552000"},
		"X-Content-Type-Options":    {Enabled: true, Value: "nosniff"},
		"Referrer-Policy":           {Enabled: true, Value: "strict-origin-when-cross-origin"},
		"X-Frame-Options":           {Enabled: true, Value: "SAMEORIGIN"},
	}
}

// HeaderCatalog returns known instance headers for the Policy UI (CSP excluded).
func HeaderCatalog() []string {
	return []string{
		"Strict-Transport-Security",
		"X-Content-Type-Options",
		"Referrer-Policy",
		"X-Frame-Options",
		"Permissions-Policy",
	}
}

const strictPermissionsPolicy = "accelerometer=(), camera=(), geolocation=(), gyroscope=(), microphone=()"

// ResolvedHeaders returns response header map for Caddy (enabled headers only).
func ResolvedHeaders(inst Instance, site Site) map[string]string {
	prof := EffectiveProfile(inst, site)
	out := make(map[string]string)
	for name, ent := range inst.Headers {
		if !ent.Enabled || ent.Value == "" {
			continue
		}
		out[name] = ent.Value
	}
	switch prof {
	case ProfileStrict:
		out["Strict-Transport-Security"] = "max-age=31536000"
		out["Permissions-Policy"] = strictPermissionsPolicy
	case ProfileBaseline:
		if _, ok := inst.Headers["Strict-Transport-Security"]; ok {
			if ent := inst.Headers["Strict-Transport-Security"]; ent.Enabled {
				out["Strict-Transport-Security"] = ent.Value
			}
		} else {
			out["Strict-Transport-Security"] = "max-age=15552000"
		}
		delete(out, "Permissions-Policy")
	}
	return out
}
