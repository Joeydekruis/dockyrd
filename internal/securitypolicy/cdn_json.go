package securitypolicy

import (
	"encoding/json"
	"strings"
)

// ParseCDNRangeJSON decodes cdn_ranges_json.
func ParseCDNRangeJSON(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}
