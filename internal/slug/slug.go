package slug

import (
	"regexp"
	"strings"
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// FromName turns a display name into a URL-safe slug base.
func FromName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonSlug.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "project"
	}
	return s
}
