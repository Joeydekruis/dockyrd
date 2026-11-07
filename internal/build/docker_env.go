package build

import (
	"os"
	"path/filepath"
)

const defaultDataDir = "/var/lib/dockyrd"

var dockerCLIPluginNames = []string{
	"docker-buildx",
	"docker-compose",
}

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
	ensureDockerCLIPlugins(dockerConfig)
	return []string{
		"DOCKER_CONFIG=" + dockerConfig,
		"BUILDX_CONFIG=" + buildxConfig,
	}
}

// ensureDockerCLIPlugins symlinks buildx/compose plugins into DOCKER_CONFIG when missing.
// Docker Desktop on macOS only installs plugins under ~/.docker/cli-plugins; a custom
// DOCKER_CONFIG path would otherwise hide them and break "docker buildx".
func ensureDockerCLIPlugins(dockerConfig string) {
	pluginsDir := filepath.Join(dockerConfig, "cli-plugins")
	_ = os.MkdirAll(pluginsDir, 0o750)
	for _, name := range dockerCLIPluginNames {
		dst := filepath.Join(pluginsDir, name)
		if _, err := os.Lstat(dst); err == nil {
			continue
		}
		for _, src := range dockerCLIPluginSearchPaths(name) {
			if _, err := os.Stat(src); err != nil {
				continue
			}
			_ = os.Symlink(src, dst)
			break
		}
	}
}

func dockerCLIPluginSearchPaths(plugin string) []string {
	var paths []string
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".docker", "cli-plugins", plugin))
	}
	paths = append(paths,
		filepath.Join("/usr/libexec/docker/cli-plugins", plugin),
		filepath.Join("/usr/local/lib/docker/cli-plugins", plugin),
	)
	return paths
}
