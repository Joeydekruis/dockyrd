package server

// buildVersion is set from cmd/dockyard via SetBuildVersion.
var buildVersion = "dev"

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
