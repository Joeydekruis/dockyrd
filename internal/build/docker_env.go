package build

import (
	"os"
	"path/filepath"
)

const defaultDataDir = "/var/lib/dockyrd"

// DockerCLIEnv returns env vars that redirect docker/buildx config into a writable data dir.
// Required when the panel runs under systemd ProtectHome (default), which makes /root read-only.
func DockerCLIEnv() []string {
	dataDir := defaultDataDir
	if v := os.Getenv("dockyrd_DATA_DIR"); v != "" {
		dataDir = v
	}
	dockerConfig := filepath.Join(dataDir, "docker-cli")
	buildxConfig := filepath.Join(dataDir, "buildx")
	_ = os.MkdirAll(dockerConfig, 0o750)
	_ = os.MkdirAll(buildxConfig, 0o750)
	return []string{
		"DOCKER_CONFIG=" + dockerConfig,
		"BUILDX_CONFIG=" + buildxConfig,
	}
}
