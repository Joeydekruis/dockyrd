package refs

import (
	"fmt"
	"regexp"
	"strings"
)

// Catalog maps service slug → published variable name → value.
type Catalog map[string]map[string]string

var refPattern = regexp.MustCompile(`\$\{\{([a-z0-9][a-z0-9-]*)\.([A-Z][A-Z0-9_]*)\}\}`)

type refKey struct {
	slug string
	key  string
}

const maxResolvePasses = 32

// Resolve substitutes all ${{slug.VAR}} placeholders using published vars only.
func Resolve(value string, catalog Catalog) (string, error) {
	if value == "" || !strings.Contains(value, "${{") {
		return value, nil
	}
	return expand(value, catalog, nil)
}

func expand(value string, catalog Catalog, stack []refKey) (string, error) {
	out := value
	for pass := 0; pass < maxResolvePasses; pass++ {
		loc := refPattern.FindStringSubmatchIndex(out)
		if loc == nil {
			return out, nil
		}
		slug := out[loc[2]:loc[3]]
		key := out[loc[4]:loc[5]]
		rk := refKey{slug: slug, key: key}
		for _, seen := range stack {
			if seen == rk {
				return "", fmt.Errorf("cyclic reference: %s.%s", slug, key)
			}
		}
		vars, ok := catalog[slug]
		if !ok {
			return "", fmt.Errorf("unknown service slug %q in reference", slug)
		}
		replacement, ok := vars[key]
		if !ok {
			return "", fmt.Errorf("unknown variable %q for service %q", key, slug)
		}
		inner, err := expand(replacement, catalog, append(stack, rk))
		if err != nil {
			return "", err
		}
		out = out[:loc[0]] + inner + out[loc[1]:]
	}
	return "", fmt.Errorf("reference resolution exceeded max passes")
}

// ContainsRef reports whether s has a reference placeholder.
func ContainsRef(s string) bool {
	return refPattern.MatchString(s)
}
