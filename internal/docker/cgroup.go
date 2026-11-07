package docker

import (
	"github.com/docker/docker/api/types/container"
)

// Cgroup placement for dockyrd-managed containers (build resilience Layer 1).
const (
	cgroupParentSystem = "dy-system.slice"
	cgroupParentApps   = "dy-apps.slice"
	cgroupParentBuild  = "dy-build.slice"
)

func applySystemCgroup(h *container.HostConfig) {
	if h == nil {
		return
	}
	h.Resources.CgroupParent = cgroupParentSystem
	h.OomScoreAdj = -900
}

func applyAppsCgroup(h *container.HostConfig) {
	if h == nil {
		return
	}
	h.Resources.CgroupParent = cgroupParentApps
	h.OomScoreAdj = 0
}

func applyBuildCgroup(h *container.HostConfig) {
	if h == nil {
		return
	}
	h.Resources.CgroupParent = cgroupParentBuild
	h.OomScoreAdj = 1000
}
