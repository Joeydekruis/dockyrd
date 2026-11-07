package deploy

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/build"
)

// resolveBuildx picks a docker buildx builder from buildLocation (local | auto | remote:<id>).
func (rt *Runtime) resolveBuildx(ctx context.Context, buildLocation string) (buildxName, metricsLabel string, remote bool, err error) {
	loc := strings.TrimSpace(strings.ToLower(buildLocation))
	if loc == "" || loc == "local" {
		if rt.Docker != nil {
			if err := rt.Docker.EnsureBuildkit(ctx, rt.Cfg.DataDir); err != nil {
				return "", "", false, fmt.Errorf("ensure buildkit: %w", err)
			}
		}
		return build.LocalBuildxBuilder(), "local", false, nil
	}
	if loc == "github_actions" {
		return "", "", false, fmt.Errorf("build runs on GitHub Actions")
	}
	if loc == "auto" {
		if name, label, ok := rt.firstOnlineRemoteBuildx(ctx); ok {
			return name, label, true, nil
		}
		if rt.Docker != nil {
			if err := rt.Docker.EnsureBuildkit(ctx, rt.Cfg.DataDir); err != nil {
				return "", "", false, fmt.Errorf("ensure buildkit: %w", err)
			}
		}
		return build.LocalBuildxBuilder(), "local", false, nil
	}
	if strings.HasPrefix(loc, "remote:") {
		id := strings.TrimSpace(strings.TrimPrefix(loc, "remote:"))
		if id == "" {
			return "", "", false, fmt.Errorf("remote builder id required")
		}
		name, err := rt.ensureRemoteBuildxByID(ctx, id)
		if err != nil {
			return "", "", false, err
		}
		return name, "remote:" + id, true, nil
	}
	return "", "", false, fmt.Errorf("unknown build location %q", buildLocation)
}

func (rt *Runtime) firstOnlineRemoteBuildx(ctx context.Context) (buildxName, label string, ok bool) {
	if rt.Q == nil {
		return "", "", false
	}
	rows, err := rt.Q.ListBuilders(ctx)
	if err != nil {
		return "", "", false
	}
	for _, row := range rows {
		if !row.Status.Valid || row.Status.String != "online" {
			continue
		}
		name, err := rt.ensureRemoteBuildxByID(ctx, row.ID)
		if err != nil {
			continue
		}
		return name, "remote:" + row.ID, true
	}
	return "", "", false
}

func (rt *Runtime) ensureRemoteBuildxByID(ctx context.Context, id string) (string, error) {
	if rt.Q == nil {
		return "", fmt.Errorf("database unavailable")
	}
	row, err := rt.Q.GetBuilder(ctx, id)
	if err != nil {
		return "", fmt.Errorf("builder not found")
	}
	name := "dy-remote-" + id
	certDir := filepath.Join(rt.Cfg.DataDir, "builders", id)
	if err := build.EnsureRemoteBuildx(ctx, name, row.Endpoint, certDir); err != nil {
		return "", err
	}
	return name, nil
}
