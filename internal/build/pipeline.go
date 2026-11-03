package build

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
)

// GitAppBuildInput runs clone → builder → image tag for app deploys.
type GitAppBuildInput struct {
	WorkDir      string
	DataDir      string
	MinFreeDiskPercent int
	DeploymentID string
	ServiceID    string
	App          appconfig.AppConfig
	Tag          string
	CloneURL     string
	BuildArgs    map[string]string
	ScrubSecrets []string
}

// RunGitAppBuild clones, builds with dockerfile or railpack, and removes the source tree.
func RunGitAppBuild(ctx context.Context, in GitAppBuildInput) (logPath string, err error) {
	if in.DataDir != "" {
		if err := CheckDiskGuard(in.DataDir, in.MinFreeDiskPercent); err != nil {
			return "", err
		}
	}
	base := filepath.Join(in.WorkDir, in.DeploymentID)
	srcDir := filepath.Join(base, "src")
	logPath = filepath.Join(base, "build.log")
	defer func() { _ = os.RemoveAll(srcDir) }()

	url, branch, err := in.App.Source.CloneRemote()
	if err != nil {
		return logPath, err
	}
	if in.CloneURL != "" {
		url = in.CloneURL
	}
	if err := CloneShallow(ctx, url, branch, srcDir); err != nil {
		return logPath, err
	}
	contextDir := in.App.BuildContextDir(srcDir)
	builder := in.App.EffectiveBuilder(contextDir)

	switch builder {
	case appconfig.BuilderDockerfile:
		err = BuildDockerfile(ctx, DockerfileInput{
			ContextDir:   contextDir,
			Dockerfile:   in.App.DockerfilePath(),
			Tag:          in.Tag,
			LogPath:      logPath,
			BuildArgs:    in.BuildArgs,
			ScrubSecrets: in.ScrubSecrets,
		})
	case appconfig.BuilderRailpack:
		err = BuildRailpack(ctx, RailpackInput{
			ContextDir:   contextDir,
			Tag:          in.Tag,
			LogPath:      logPath,
			ServiceID:    in.ServiceID,
			BuildArgs:    in.BuildArgs,
			ScrubSecrets: in.ScrubSecrets,
		})
	default:
		err = errUnsupportedBuilder(builder)
	}
	return logPath, err
}

// DeploymentWorkDir returns the builds root for one deployment.
func DeploymentWorkDir(buildsRoot, deploymentID string) string {
	return filepath.Join(buildsRoot, deploymentID)
}
