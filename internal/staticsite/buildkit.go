package staticsite

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/build"
)

// exportWithBuildkit runs buildx local export from an existing source tree.
func exportWithBuildkit(ctx context.Context, dockerBin string, in BuildInput, workDir, shellCmd string, buildEnv []string) (outputDir string, err error) {
	if dockerBin == "" {
		dockerBin = "docker"
	}
	base := filepath.Join(in.WorkDir, in.DeploymentID)
	logPath := in.LogPath
	if logPath == "" {
		logPath = filepath.Join(base, "build.log")
	}
	exportDir := filepath.Join(base, "export")
	if err := os.RemoveAll(exportDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o750); err != nil {
		return "", err
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return "", err
	}
	defer logFile.Close()

	copyRel := staticOutputCopyRel(in.OutputDir)
	dfPath := filepath.Join(workDir, "Dockerfile.dockyrd")
	if err := os.WriteFile(dfPath, []byte(renderStaticDockerfile(in.NodeImage, shellCmd, copyRel, buildEnv)), 0o644); err != nil {
		return "", err
	}

	args := []string{"buildx", "build"}
	args = build.AppendBuilderArgs(ctx, args)
	args = append(args,
		"-f", dfPath,
		"--output", "type=local,dest="+exportDir,
		workDir,
	)
	cmd := exec.CommandContext(ctx, dockerBin, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("static buildkit: %w", err)
	}
	return DetectOutputDir(exportDir, in.OutputDir)
}

func staticOutputCopyRel(explicit string) string {
	explicit = strings.TrimSpace(explicit)
	if explicit == "" || explicit == "." {
		return "."
	}
	return strings.Trim(explicit, "/\\")
}

func renderStaticDockerfile(nodeImage, shellCmd, copyRel string, env []string) string {
	envLines := ""
	for _, e := range env {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		envLines += fmt.Sprintf("ENV %s\n", e)
	}
	copySrc := "/src"
	if copyRel != "." {
		copySrc = "/src/" + copyRel
	}
	return fmt.Sprintf(`# syntax=docker/dockerfile:1
FROM %s AS build
WORKDIR /src
COPY . .
%sRUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    --mount=type=cache,target=/root/.npm \
    sh -lc %q

FROM scratch
COPY --from=build %s /
`, nodeImage, envLines, shellCmd, copySrc)
}
