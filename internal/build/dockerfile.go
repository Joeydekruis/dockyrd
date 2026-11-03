package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DockerfileInput configures a local docker buildx build --load.
type DockerfileInput struct {
	ContextDir   string
	Dockerfile   string
	Tag          string
	LogPath      string
	BuildArgs    map[string]string
	ScrubSecrets []string
	DockerBinary string
}

// BuildDockerfile runs docker buildx build --load and streams output to LogPath.
func BuildDockerfile(ctx context.Context, in DockerfileInput) error {
	if in.DockerBinary == "" {
		in.DockerBinary = "docker"
	}
	contextDir, err := filepath.Abs(in.ContextDir)
	if err != nil {
		return err
	}
	df := strings.TrimSpace(in.Dockerfile)
	if df == "" {
		df = "Dockerfile"
	}
	tag := strings.TrimSpace(in.Tag)
	if tag == "" {
		return fmt.Errorf("image tag required")
	}
	if err := os.MkdirAll(filepath.Dir(in.LogPath), 0o750); err != nil {
		return err
	}
	logFile, err := os.Create(in.LogPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	secrets := in.ScrubSecrets
	if len(secrets) == 0 {
		for _, v := range in.BuildArgs {
			secrets = append(secrets, v)
		}
	}
	writer := ScrubWriter{W: logFile, Secrets: secrets}

	args := []string{
		"buildx", "build",
		"--load",
	}
	args = appendBuilderArgs(args)
	args = append(args,
		"-f", df,
		"-t", tag,
	)
	for k, v := range in.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, contextDir)

	cmd := exec.CommandContext(ctx, in.DockerBinary, args...)
	cmd.Stdout = writer
	cmd.Stderr = writer
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build: %w", err)
	}
	return nil
}
