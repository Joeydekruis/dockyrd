package server

import "time"

// buildVersion is set from cmd/dockyard via SetBuildVersion.
var buildVersion = "dev"

var processStartedAt = time.Now().Unix()

// SetBuildVersion records the linker-injected release version.
func SetBuildVersion(v string) {
	if v != "" {
		buildVersion = v
	}
}

// BuildVersion returns the running binary version string.
func BuildVersion() string {
	return buildVersion
}

// ProcessStartedAt is when the current process began serving (Unix seconds).
func ProcessStartedAt() int64 {
	return processStartedAt
}
