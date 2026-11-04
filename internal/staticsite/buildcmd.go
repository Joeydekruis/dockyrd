package staticsite

import "strings"

const legacyDefaultBuildCmd = "npm ci && npm run build"

// ResolveBuildCommand returns the shell command to run in the build container.
// An empty string means skip the container and publish from the clone directly.
func ResolveBuildCommand(cfg Config, workDir string) string {
	if cfg.BuildEnabled != nil && !*cfg.BuildEnabled {
		return ""
	}

	i := sanitizeCmdPart(cfg.InstallCmd)
	b := sanitizeCmdPart(cfg.BuildScript)
	if i != "" || b != "" {
		if i == "" {
			return b
		}
		if b == "" {
			return i
		}
		return i + " && " + b
	}

	legacy := sanitizeCmdPart(cfg.BuildCmd)
	if legacy == "" {
		return ""
	}
	if legacy == legacyDefaultBuildCmd {
		return DefaultBuildCmd(workDir, legacy)
	}
	return legacy
}

func sanitizeCmdPart(s string) string {
	s = strings.TrimSpace(s)
	if s == "false" {
		return ""
	}
	return s
}
