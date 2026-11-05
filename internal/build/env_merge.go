package build

import "strings"

// MergeBuildArgs adds injected env defaults without overriding user-provided keys.
func MergeBuildArgs(existing map[string]string, ramBudget, swapBudget int64) map[string]string {
	out := make(map[string]string, len(existing)+6)
	for k, v := range existing {
		out[k] = v
	}
	var prior []string
	for k, v := range out {
		prior = append(prior, k+"="+v)
	}
	for _, e := range InjectBuildEnv(prior, ramBudget, swapBudget) {
		k, v, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}
		if _, exists := out[k]; !exists {
			out[k] = v
		}
	}
	return out
}
