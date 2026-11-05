package netsafe

import (
	"strings"
)

// RedactError replaces secretURL substrings in err text (e.g. Slack webhook tokens in *url.Error).
func RedactError(err error, secretURL string) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	if secretURL != "" {
		s = strings.ReplaceAll(s, secretURL, "••••")
	}
	return s
}
