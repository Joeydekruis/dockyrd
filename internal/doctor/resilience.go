package doctor

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/resources"
)

func checkBuildResilience(ctx context.Context, cfg config.Config) []Check {
	if runtime.GOOS != "linux" {
		return []Check{{
			Name:   "Build isolation (cgroup v2)",
			Status: StatusWarn,
			Detail: "not applicable on " + runtime.GOOS,
		}}
	}
	var out []Check
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err != nil {
		out = append(out, Check{
			Name:    "cgroup v2",
			Status:  StatusFail,
			Detail:  "unified hierarchy missing",
			FixHint: "Use Ubuntu 22.04+, Debian 12+, or reinstall with cgroup v2 enabled",
		})
	} else {
		out = append(out, Check{Name: "cgroup v2", Status: StatusOK, Detail: "unified hierarchy"})
	}
	for _, unit := range []string{"dy.slice", "dy-system.slice", "dy-apps.slice", "dy-build.slice"} {
		path := filepath.Join("/etc/systemd/system", unit)
		if _, err := os.Stat(path); err != nil {
			out = append(out, Check{
				Name:    "Slice " + unit,
				Status:  StatusWarn,
				Detail:  "missing",
				FixHint: "Re-run install.sh or copy packaging/" + unit,
			})
		} else {
			out = append(out, Check{Name: "Slice " + unit, Status: StatusOK})
		}
	}
	res := resources.New(cfg.BuildIsolationEnabled())
	if !res.Available() {
		out = append(out, Check{
			Name:   "Build resource driver",
			Status: StatusWarn,
			Detail: "inactive (dockyrd_BUILD_ISOLATION=0 or no cgroup v2)",
		})
	} else {
		freezePath := filepath.Join(resources.SliceBuildPath(), "cgroup.freeze")
		b, err := os.ReadFile(freezePath)
		if err != nil {
			out = append(out, Check{Name: "Build slice freeze", Status: StatusWarn, Detail: err.Error()})
		} else if strings.TrimSpace(string(b)) != "0" {
			out = append(out, Check{
				Name:    "Build slice freeze",
				Status:  StatusFail,
				Detail:  "frozen while no build should run",
				FixHint: "Restart dockyrd.service to thaw dy-build.slice",
			})
		} else {
			out = append(out, Check{Name: "Build slice freeze", Status: StatusOK, Detail: "thawed"})
		}
	}
	toml := filepath.Join(docker.BuildkitConfigDir(cfg.DataDir), "buildkitd.toml")
	if _, err := os.Stat(toml); err != nil {
		out = append(out, Check{
			Name:    "BuildKit config",
			Status:  StatusWarn,
			Detail:  "not rendered yet",
			FixHint: "Start dockyrd serve (EnsureBuildkit renders buildkitd.toml)",
		})
	} else {
		out = append(out, Check{Name: "BuildKit config", Status: StatusOK, Detail: toml})
	}
	if b, err := os.ReadFile("/proc/swaps"); err == nil && len(strings.TrimSpace(string(b))) <= 1 {
		out = append(out, Check{
			Name:   "Swap",
			Status: StatusWarn,
			Detail: "none configured",
			FixHint: "Add swap for small servers (install.sh can create a swapfile)",
		})
	} else {
		out = append(out, Check{Name: "Swap", Status: StatusOK, Detail: "present"})
	}
	_ = ctx
	return out
}
