package proxybuild

// PanelVersion is the running dockyrd semver (set from cmd on serve).
var PanelVersion = "dev"

// SetPanelVersion records the linker-injected panel version for manifest resolution.
func SetPanelVersion(v string) {
	if v != "" {
		PanelVersion = v
	}
}