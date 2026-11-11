package mcp

import "strings"

// OAuth scopes for Dockyrd MCP (ADR-0002).
const (
	ScopeRead   = "mcp:read"
	ScopeDeploy = "mcp:deploy"
	ScopeWrite  = "mcp:write"
	ScopeInstRO = "mcp:instance:read"
)

// AllScopes lists scopes advertised in PRM / authorization server metadata.
var AllScopes = []string{ScopeRead, ScopeDeploy, ScopeWrite, ScopeInstRO}

// ValidScope reports whether s is a known MCP scope.
func ValidScope(s string) bool {
	for _, v := range AllScopes {
		if v == s {
			return true
		}
	}
	return false
}

// ParseScopeList splits a space-separated scope string and drops unknown entries.
func ParseScopeList(raw string) []string {
	var out []string
	for _, part := range strings.Fields(raw) {
		if ValidScope(part) {
			out = append(out, part)
		}
	}
	return out
}
