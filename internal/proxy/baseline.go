package proxy

// Baseline edge defaults (SECURITY_BRIEFING §3.2 Baseline) until instance policy (P1c).

func baselineHeadersHandler() map[string]any {
	return map[string]any{
		"handler": "headers",
		"response": map[string]any{
			"set": map[string]any{
				"Strict-Transport-Security": []string{"max-age=15552000"},
				"X-Content-Type-Options":    []string{"nosniff"},
				"Referrer-Policy":           []string{"strict-origin-when-cross-origin"},
				"X-Frame-Options":           []string{"SAMEORIGIN"},
			},
			"deferred": true,
			"delete":   []string{"Server"},
		},
	}
}

func encodeHandler() map[string]any {
	return map[string]any{
		"handler": "encode",
		"encodings": map[string]any{
			"zstd": map[string]any{},
			"gzip": map[string]any{},
		},
		"prefer": []string{"zstd", "gzip"},
	}
}

func prependBaseline(handles []any) []any {
	out := make([]any, 0, len(handles)+2)
	out = append(out, baselineHeadersHandler(), encodeHandler())
	return append(out, handles...)
}

func staticImmutableCacheHandler() map[string]any {
	return map[string]any{
		"handler": "headers",
		"match": map[string]any{
			"path": []string{
				"/_next/static/*",
				"*.css",
				"*.js",
				"*.woff2",
			},
		},
		"response": map[string]any{
			"set": map[string]any{
				"Cache-Control": []string{"public, max-age=31536000, immutable"},
			},
		},
	}
}

func staticHTMLNoCacheHandler() map[string]any {
	return map[string]any{
		"handler": "headers",
		"match": map[string]any{
			"path": []string{"/", "/index.html", "*.html"},
		},
		"response": map[string]any{
			"set": map[string]any{
				"Cache-Control": []string{"no-cache"},
			},
		},
	}
}

func serverTimeouts() map[string]any {
	return map[string]any{
		"read_header_timeout": "10s",
		"idle_timeout":        "2m",
		"max_header_bytes":    1_048_576,
	}
}

func mergeServerConfig(base map[string]any, trusted TrustedProxyConfig) map[string]any {
	for k, v := range serverTimeouts() {
		base[k] = v
	}
	// HTTP/3 needs UDP 443; many cloud firewalls (Hetzner) only open TCP — Chrome then shows ERR_QUIC_PROTOCOL_ERROR.
	base["protocols"] = []string{"h1", "h2"}
	applyTrustedProxies(base, trusted)
	return base
}
