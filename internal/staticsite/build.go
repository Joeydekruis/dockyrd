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

// BuildInput runs install/build inside a disposable Node container.
type BuildInput struct {
	CloneURL     string
	Branch       string
	RootDir      string
	BuildCmd     string
	NodeImage    string
	WorkDir      string
	DeploymentID string
	LogPath      string
	OutputDir    string
}

// Build clones and runs the build command in Docker.
func Build(ctx context.Context, dockerBin string, in BuildInput) (outputDir string, err error) {
	if dockerBin == "" {
		dockerBin = "docker"
	}
	base := filepath.Join(in.WorkDir, in.DeploymentID)
	srcDir := filepath.Join(base, "src")
	logPath := in.LogPath
	if logPath == "" {
		logPath = filepath.Join(base, "build.log")
	}
	defer func() { _ = os.RemoveAll(srcDir) }()

	if err := build.CloneShallow(ctx, in.CloneURL, in.Branch, srcDir); err != nil {
		return "", err
	}
	workDir := srcDir
	if root := strings.TrimSpace(in.RootDir); root != "" && root != "." {
		workDir = filepath.Join(srcDir, root)
	}
	buildCmd := DefaultBuildCmd(workDir, in.BuildCmd)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o750); err != nil {
		return "", err
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return "", err
	}
	defer logFile.Close()

	args := []string{
		"run", "--rm",
		"-v", workDir + ":/src",
		"-w", "/src",
		in.NodeImage,
		"sh", "-lc", buildCmd,
	}
	cmd := exec.CommandContext(ctx, dockerBin, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("static build: %w", err)
	}
	return DetectOutputDir(workDir, in.OutputDir)
}
