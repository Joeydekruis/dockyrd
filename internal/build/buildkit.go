package build

import (
	"os"
	"strings"
)

const defaultBuildxBuilder = "dy-buildkit"

// BuildxBuilder returns optional --builder name (env DOCKYARD_BUILDX_BUILDER or dy-buildkit).
func BuildxBuilder() string {
	if v := strings.TrimSpace(os.Getenv("DOCKYARD_BUILDX_BUILDER")); v != "" {
		return v
	}
	return defaultBuildxBuilder
}

func appendBuilderArgs(args []string) []string {
	if b := BuildxBuilder(); b != "" {
		args = append(args, "--builder", b)
	}
	return args
}
