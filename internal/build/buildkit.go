package build

import (
	"context"
	"os"
	"strings"
)

const defaultBuildxBuilder = "dy-buildkit"

type buildxBuilderKey struct{}

// WithBuildxBuilder attaches a docker buildx --builder name to ctx for nested build commands.
func WithBuildxBuilder(ctx context.Context, name string) context.Context {
	name = strings.TrimSpace(name)
	if name == "" {
		return ctx
	}
	return context.WithValue(ctx, buildxBuilderKey{}, name)
}

// LocalBuildxBuilder is the on-host BuildKit builder (dy-buildkit).
func LocalBuildxBuilder() string {
	return defaultBuildxBuilder
}

// BuildxBuilder returns optional --builder name (env dockyrd_BUILDX_BUILDER or dy-buildkit).
func BuildxBuilder() string {
	if v := strings.TrimSpace(os.Getenv("dockyrd_BUILDX_BUILDER")); v != "" {
		return v
	}
	return defaultBuildxBuilder
}

func builderFromContext(ctx context.Context) string {
	if ctx != nil {
		if v, ok := ctx.Value(buildxBuilderKey{}).(string); ok {
			v = strings.TrimSpace(v)
			if v != "" {
				return v
			}
		}
	}
	return BuildxBuilder()
}

func appendBuilderArgs(ctx context.Context, args []string) []string {
	return AppendBuilderArgs(ctx, args)
}

// AppendBuilderArgs adds --builder from ctx or env default.
func AppendBuilderArgs(ctx context.Context, args []string) []string {
	if b := builderFromContext(ctx); b != "" {
		args = append(args, "--builder", b)
	}
	return args
}

// AppendBuilderArgsPublic is deprecated; use AppendBuilderArgs with context.
func AppendBuilderArgsPublic(args []string) []string {
	return AppendBuilderArgs(context.Background(), args)
}
