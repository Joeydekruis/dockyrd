package updater

import (
	"strconv"
	"strings"
)

// CompareVersions compares semver-like strings (optional leading v). Returns -1, 0, or 1.
func CompareVersions(current, latest string) int {
	c := parseParts(current)
	l := parseParts(latest)
	for i := 0; i < 3; i++ {
		if c[i] < l[i] {
			return -1
		}
		if c[i] > l[i] {
			return 1
		}
	}
	return 0
}

func parseParts(v string) [3]int {
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	parts := strings.SplitN(v, "-", 2)[0]
	segs := strings.Split(parts, ".")
	var out [3]int
	for i := 0; i < 3 && i < len(segs); i++ {
		n, _ := strconv.Atoi(segs[i])
		out[i] = n
	}
	return out
}

func normalizeTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return tag
	}
	if !strings.HasPrefix(tag, "v") {
		return "v" + tag
	}
	return tag
}
