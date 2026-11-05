package resources

// Systemd slice names for Dockyard cgroup isolation (Layer 1).
const (
	SliceRoot   = "dy.slice"
	SliceSystem = "dy-system.slice"
	SliceApps   = "dy-apps.slice"
	SliceBuild  = "dy-build.slice"
)

// SliceBuildPath returns the cgroup path for the build slice.
func SliceBuildPath() string {
	return sliceCgroupPath(SliceBuild)
}

// OOM score adjustments per briefing.
const (
	OOMScoreSystem = -900
	OOMScoreApps   = 0
	OOMScoreBuild  = 1000
)
