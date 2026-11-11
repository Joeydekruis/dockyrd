package mcp

import (
	"net/url"
	"strings"
)

// SafeReturnPath returns path+query when returnTo is a same-origin relative path.
func SafeReturnPath(returnTo string) string {
	returnTo = strings.TrimSpace(returnTo)
	if returnTo == "" {
		return ""
	}
	if strings.HasPrefix(returnTo, "//") {
		return ""
	}
	if strings.Contains(returnTo, "://") {
		u, err := url.Parse(returnTo)
		if err != nil || u.Scheme != "" {
			return ""
		}
	}
	if !strings.HasPrefix(returnTo, "/") {
		return ""
	}
	if strings.HasPrefix(returnTo, "/sign-in") || strings.HasPrefix(returnTo, "/setup") {
		return ""
	}
	return returnTo
}
