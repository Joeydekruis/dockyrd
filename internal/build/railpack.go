package build

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RailpackInput configures railpack prepare + buildx load.
type RailpackInput struct {
	ContextDir   string
	Tag          string
	LogPath      string
	ServiceID    string
	BuildArgs    map[string]string
	ScrubSecrets []string
	RailpackBin  string
	DockerBinary string
	FrontendRef  string
}

// BuildRailpack runs railpack prepare then docker buildx using plan.json.
func BuildRailpack(ctx context.Context, in RailpackInput) error {
	if in.RailpackBin == "" {
		in.RailpackBin = "railpack"
	}
	if in.DockerBinary == "" {
		in.DockerBinary = "docker"
	}
	if in.FrontendRef == "" {
		in.FrontendRef = DefaultRailpackFrontend
	}
	contextDir, err := filepath.Abs(in.ContextDir)
	if err != nil {
		return err
	}
	planPath := filepath.Join(filepath.Dir(in.LogPath), "plan.json")
	infoPath := filepath.Join(filepath.Dir(in.LogPath), "info.json")
	if err := os.MkdirAll(filepath.Dir(in.LogPath), 0o750); err != nil {
		return err
	}
	logFile, err := os.Create(in.LogPath)
	if err != nil {
		return err
	}
	defer logFile.Close()

	secrets := in.ScrubSecrets
	writer := ScrubWriter{W: logFile, Secrets: secrets}

	prepareArgs := []string{
		"prepare", contextDir,
		"--plan-out", planPath,
		"--info-out", infoPath,
	}
	for k, v := range in.BuildArgs {
		prepareArgs = append(prepareArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	const maxPrepareAttempts = 3
	var prepareErr error
	for attempt := 0; attempt < maxPrepareAttempts; attempt++ {
		cmd := exec.CommandContext(ctx, in.RailpackBin, prepareArgs...)
		cmd.Stdout = writer
		cmd.Stderr = writer
		prepareErr = cmd.Run()
		if prepareErr == nil {
			break
		}
		if exitErr, ok := prepareErr.(*exec.ExitError); ok && exitErr.ExitCode() == 75 && attempt+1 < maxPrepareAttempts {
			continue
		}
		if msg := railpackInfoMessage(infoPath); msg != "" {
			return fmt.Errorf("railpack prepare: %s", msg)
		}
		return fmt.Errorf("railpack prepare: %w", prepareErr)
	}

	buildWriter := ScrubWriter{W: logFile, Secrets: secrets}
	buildArgs := []string{
		"buildx", "build",
		"--load",
	}
	buildArgs = appendBuilderArgs(ctx, buildArgs)
	buildArgs = append(buildArgs,
		"--build-arg", "BUILDKIT_SYNTAX="+in.FrontendRef,
		"--build-arg", "cache-key="+in.ServiceID,
		"-f", planPath,
		"-t", in.Tag,
	)
	for k := range in.BuildArgs {
		buildArgs = append(buildArgs, "--secret", fmt.Sprintf("id=%s,env=%s", k, k))
	}
	buildArgs = append(buildArgs, contextDir)

	buildCmd := exec.CommandContext(ctx, in.DockerBinary, buildArgs...)
	buildCmd.Env = append(os.Environ(), DockerCLIEnv()...)
	buildCmd.Env = append(buildCmd.Env, envPairs(in.BuildArgs)...)
	buildCmd.Stdout = buildWriter
	buildCmd.Stderr = buildWriter
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("railpack build: %w", err)
	}
	return nil
}

func envPairs(args map[string]string) []string {
	out := make([]string, 0, len(args))
	for k, v := range args {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return out
}

func railpackInfoMessage(infoPath string) string {
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return ""
	}
	var doc struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return strings.TrimSpace(string(data))
	}
	for _, e := range doc.Errors {
		if strings.TrimSpace(e.Message) != "" {
			return e.Message
		}
	}
	return ""
}
