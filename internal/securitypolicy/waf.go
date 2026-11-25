package securitypolicy

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// WAFMode values match service_security.waf_mode CHECK.
type WAFMode string

const (
	WAFInherit WAFMode = "inherit"
	WAFOff     WAFMode = "off"
	WAFDetect  WAFMode = "detect"
	WAFBlock   WAFMode = "block"
)

const (
	maxWAFPathExceptions = 64
	maxWAFPathLen        = 256
)

// EffectiveWAFMode resolves stored waf_mode against the effective profile.
// Strict profile never auto-blocks: inherit under Strict → detect only.
// Custom with inherit uses the Baseline floor (off).
func EffectiveWAFMode(inst Instance, site Site) WAFMode {
	stored := WAFMode(strings.TrimSpace(string(site.WAFMode)))
	if stored == "" {
		stored = WAFInherit
	}
	switch stored {
	case WAFOff, WAFDetect, WAFBlock:
		return stored
	case WAFInherit:
		switch EffectiveProfile(inst, site) {
		case ProfileStrict:
			return WAFDetect
		default:
			// Baseline, Custom, unknown → off floor (never auto-block).
			return WAFOff
		}
	default:
		return WAFOff
	}
}

// ValidateWAFMode checks a stored waf_mode value.
func ValidateWAFMode(mode string) error {
	switch WAFMode(strings.TrimSpace(mode)) {
	case WAFInherit, WAFOff, WAFDetect, WAFBlock:
		return nil
	default:
		return fmt.Errorf("invalid waf_mode %q", mode)
	}
}

// ValidatePathExceptions normalizes path prefixes for WAF exceptions.
func ValidatePathExceptions(paths []string) ([]string, error) {
	if len(paths) > maxWAFPathExceptions {
		return nil, fmt.Errorf("waf_path_exceptions exceeds %d entries", maxWAFPathExceptions)
	}
	out := make([]string, 0, len(paths))
	seen := make(map[string]struct{})
	for _, raw := range paths {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if utf8.RuneCountInString(raw) > maxWAFPathLen {
			return nil, fmt.Errorf("waf path exception exceeds %d characters", maxWAFPathLen)
		}
		if !strings.HasPrefix(raw, "/") {
			return nil, fmt.Errorf("waf path exception must start with /")
		}
		// Reject quote/backslash/control chars so values cannot break Coraza
		// `@beginsWith …` directives when rendered into Caddy JSON.
		if strings.Contains(raw, "..") || strings.ContainsAny(raw, "\"\\\n\r\x00") {
			return nil, fmt.Errorf("waf path exception contains invalid characters")
		}
		if _, ok := seen[raw]; ok {
			continue
		}
		seen[raw] = struct{}{}
		out = append(out, raw)
	}
	return out, nil
}
