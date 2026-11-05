package notify

import (
	"strings"
	"unicode/utf8"
)

const slackSectionMax = 2500
const slackHeaderMax = 150

func truncateRunes(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= max-1 {
			b.WriteString("…")
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

func escapeSlackMrkdwn(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func escapeDiscordMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"`", "\\`",
		"*", "\\*",
		"_", "\\_",
		"~", "\\~",
		"|", "\\|",
		">", "\\>",
	)
	return replacer.Replace(s)
}

func scrubDetail(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 500 {
		s = truncateRunes(s, 500)
	}
	return s
}

// ScrubDetailForEmit exposes detail scrubbing for cross-package tests.
func ScrubDetailForEmit(s string) string {
	return scrubDetail(s)
}

func absoluteURL(base, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if base == "" {
		return ""
	}
	if !strings.HasPrefix(href, "/") {
		href = "/" + href
	}
	return base + href
}

func levelColorHex(l Level) string {
	switch l {
	case LevelError:
		return "#B3261E"
	case LevelWarning:
		return "#F59E0B"
	case LevelSuccess:
		return "#16A34A"
	default:
		return "#64748B"
	}
}

func levelColorDiscord(l Level) int {
	switch l {
	case LevelError:
		return 11740702
	case LevelWarning:
		return 16096779
	case LevelSuccess:
		return 5763719
	default:
		return 9807270
	}
}
