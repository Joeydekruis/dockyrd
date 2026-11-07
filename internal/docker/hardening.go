package docker

import (
	"github.com/docker/docker/api/types/container"
)

const (
	defaultPidsLimit int64 = 512
	// OneCPU is Docker NanoCPUs for 1 vCPU.
	OneCPU int64 = 1_000_000_000
	oneCPU       = OneCPU
)

// containerCapsDatabase is a minimal capability set for official DB images (gosu/su-exec entrypoints).
var containerCapsDatabase = []string{
	"CHOWN",
	"DAC_OVERRIDE",
	"FOWNER",
	"SETGID",
	"SETUID",
}

// applyUserContainerHardening sets Docker security defaults for user workloads (apps + databases).
func applyUserContainerHardening(h *container.HostConfig, memLimitMB int, database bool) {
	if h == nil {
		return
	}
	if memLimitMB <= 0 {
		memLimitMB = 512
	}
	pids := defaultPidsLimit
	if h.Resources.PidsLimit != nil && *h.Resources.PidsLimit > 0 {
		pids = *h.Resources.PidsLimit
	}
	h.SecurityOpt = appendUniqueString(h.SecurityOpt, "no-new-privileges:true")
	h.CapDrop = appendUniqueString(h.CapDrop, "ALL")
	if database {
		h.CapAdd = appendUniqueCapAdd(h.CapAdd, containerCapsDatabase)
	}
	if h.Resources.PidsLimit == nil {
		h.Resources.PidsLimit = &pids
	}
	if h.Resources.NanoCPUs == 0 && !database {
		h.Resources.NanoCPUs = oneCPU
	}
	if h.Resources.Memory == 0 {
		h.Resources.Memory = int64(memLimitMB) * 1024 * 1024
	}
}

func appendUniqueString(dst []string, value string) []string {
	for _, s := range dst {
		if s == value {
			return dst
		}
	}
	return append(dst, value)
}

func appendUniqueCapAdd(dst, add []string) []string {
	seen := make(map[string]struct{}, len(dst))
	for _, c := range dst {
		seen[c] = struct{}{}
	}
	for _, c := range add {
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		dst = append(dst, c)
	}
	return dst
}
