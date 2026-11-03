package staticsite

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const defaultReleaseKeep = 5

// PruneReleases removes oldest release directories beyond keep (never deletes live path).
func PruneReleases(sitesRoot, serviceSlug string, keep int) {
	if keep <= 0 {
		keep = defaultReleaseKeep
	}
	paths := PathsForService(sitesRoot, serviceSlug)
	entries, err := os.ReadDir(paths.ReleasesDir)
	if err != nil {
		return
	}
	type item struct {
		name string
		mod  int64
	}
	var dirs []item
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		dirs = append(dirs, item{name: e.Name(), mod: info.ModTime().Unix()})
	}
	if len(dirs) <= keep {
		return
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].mod > dirs[j].mod })
	for _, d := range dirs[keep:] {
		_ = os.RemoveAll(filepath.Join(paths.ReleasesDir, d.name))
	}
}

// DefaultBuildCmd picks install/build commands from lockfiles when cfg.BuildCmd is default.
func DefaultBuildCmd(workDir, configured string) string {
	if configured != "" && configured != "npm ci && npm run build" {
		return configured
	}
	checks := []struct {
		file string
		cmd  string
	}{
		{"pnpm-lock.yaml", "corepack enable && pnpm install --frozen-lockfile && pnpm run build"},
		{"yarn.lock", "yarn install --frozen-lockfile && yarn run build"},
		{"bun.lockb", "bun install && bun run build"},
		{"package-lock.json", "npm ci && npm run build"},
	}
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(workDir, c.file)); err == nil {
			return c.cmd
		}
	}
	if _, err := os.Stat(filepath.Join(workDir, "package.json")); err == nil {
		return "npm ci && npm run build"
	}
	return strings.TrimSpace(configured)
}
