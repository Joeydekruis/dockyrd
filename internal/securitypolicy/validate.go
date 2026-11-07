package securitypolicy

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"unicode/utf8"
)

const maxCSPLen = 8 * 1024

var basicUserPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,64}$`)

// ParseHeadersJSON decodes instance headers_json.
func ParseHeadersJSON(raw string) (map[string]HeaderEntry, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultBaselineHeaders(), nil
	}
	var m map[string]HeaderEntry
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	if m == nil {
		m = map[string]HeaderEntry{}
	}
	return m, nil
}

// ValidatePolicyHeaders ensures locked minimum headers stay enabled when lock is on.
func ValidatePolicyHeaders(inst Instance) *FloorError {
	if !inst.Lock {
		return nil
	}
	required := DefaultBaselineHeaders()
	for name, req := range required {
		ent, ok := inst.Headers[name]
		if !ok || !ent.Enabled || strings.TrimSpace(ent.Value) == "" {
			return &FloorError{
				Field:   "headers_json",
				Message: fmt.Sprintf("header %q must stay enabled while policy lock is on", name),
			}
		}
		_ = req
	}
	return nil
}

// ValidateCSP checks CSP length when set.
func ValidateCSP(csp string) error {
	csp = strings.TrimSpace(csp)
	if csp == "" {
		return nil
	}
	if utf8.RuneCountInString(csp) > maxCSPLen {
		return fmt.Errorf("csp exceeds %d characters", maxCSPLen)
	}
	return nil
}

// ValidateAllowlist parses CIDRs (v4 + v6); empty list is allowed.
func ValidateAllowlist(cidrs []string) ([]string, error) {
	if len(cidrs) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(cidrs))
	seen := make(map[string]struct{})
	for _, raw := range cidrs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if _, _, err := net.ParseCIDR(raw); err != nil {
			ip := net.ParseIP(raw)
			if ip == nil {
				return nil, fmt.Errorf("invalid CIDR or IP %q", raw)
			}
			if ip.To4() != nil {
				raw = raw + "/32"
			} else {
				raw = raw + "/128"
			}
		}
		if _, _, err := net.ParseCIDR(raw); err != nil {
			return nil, fmt.Errorf("invalid CIDR %q", raw)
		}
		if _, ok := seen[raw]; ok {
			continue
		}
		seen[raw] = struct{}{}
		out = append(out, raw)
	}
	return out, nil
}

// ValidateBasicUsername validates basic auth username charset.
func ValidateBasicUsername(user string) error {
	user = strings.TrimSpace(user)
	if user == "" {
		return fmt.Errorf("username required")
	}
	if !basicUserPattern.MatchString(user) {
		return fmt.Errorf("username contains invalid characters")
	}
	return nil
}

// ValidateResources bounds mem/cpu/pids for docker deploy.
func ValidateResources(memMB *int, cpu *float64, pids *int) error {
	if memMB != nil {
		if *memMB < 64 || *memMB > 1_048_576 {
			return fmt.Errorf("mem_mb out of range")
		}
	}
	if cpu != nil {
		if *cpu <= 0 || *cpu > 256 {
			return fmt.Errorf("cpu out of range")
		}
	}
	if pids != nil {
		if *pids < 16 || *pids > 1_000_000 {
			return fmt.Errorf("pids out of range")
		}
	}
	return nil
}
