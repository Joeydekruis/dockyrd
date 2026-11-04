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
	InstallCmd   string
	BuildScript  string
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
	workDir, err := resolveStaticWorkDir(srcDir, in.RootDir, in.Branch)
	if err != nil {
		return "", err
	}
	shellCmd := ResolveBuildCommand(Config{
		InstallCmd:  in.InstallCmd,
		BuildScript: in.BuildScript,
		BuildCmd:    in.BuildCmd,
	}, workDir)
	if strings.TrimSpace(shellCmd) == "" {
		return DetectOutputDir(workDir, in.OutputDir)
	}
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
		"sh", "-lc", shellCmd,
	}
	cmd := exec.CommandContext(ctx, dockerBin, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("static build: %w", err)
	}
	return DetectOutputDir(workDir, in.OutputDir)
}

// resolveStaticWorkDir picks the build/output directory inside a fresh clone.
func resolveStaticWorkDir(cloneRoot, rootDir, branch string) (string, error) {
	root := strings.TrimSpace(rootDir)
	if root == "" || root == "." || root == "/" {
		return cloneRoot, nil
	}
	root = strings.Trim(root, "/\\")
	root = filepath.Clean(root)
	if root == ".." || strings.HasPrefix(root, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid root directory %q", rootDir)
	}
	workDir := filepath.Join(cloneRoot, root)
	info, err := os.Stat(workDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf(
				"root directory %q not found in cloned repo (branch %q). Verify repo, branch, and that the folder is pushed to GitHub",
				root,
				strings.TrimSpace(branch),
			)
		}
		return "", fmt.Errorf("root directory %q: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("root directory %q is not a directory", root)
	}
	return workDir, nil
}
