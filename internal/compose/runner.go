package compose

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Runner executes docker compose subprocesses.
type Runner interface {
	ConfigJSON(ctx context.Context, project, workDir, composeFile string, env []string) (map[string]any, error)
	Pull(ctx context.Context, project, workDir, composeFile string, env []string, log func(string)) error
	Up(ctx context.Context, project, workDir, composeFile string, env []string, waitSec int, log func(string)) error
	Start(ctx context.Context, project, workDir, composeFile string) error
	Stop(ctx context.Context, project, workDir, composeFile string) error
	Restart(ctx context.Context, project, workDir, composeFile string) error
	Down(ctx context.Context, project, workDir, composeFile string, volumes bool) error
	PS(ctx context.Context, project, workDir, composeFile string) ([]ContainerStatus, error)
}

// CLIRunner uses the docker compose CLI.
type CLIRunner struct{}

type ContainerStatus struct {
	Service string `json:"service"`
	State   string `json:"state"`
	Name    string `json:"name"`
	Health  string `json:"health,omitempty"`
}

func (r CLIRunner) ConfigJSON(ctx context.Context, project, workDir, composeFile string, env []string) (map[string]any, error) {
	out, err := r.run(ctx, project, workDir, composeFile, env, "config", "--format", "json")
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		return nil, fmt.Errorf("parse compose config json: %w", err)
	}
	return doc, nil
}

func (r CLIRunner) Pull(ctx context.Context, project, workDir, composeFile string, env []string, log func(string)) error {
	_, err := r.runLogged(ctx, project, workDir, composeFile, env, log, "pull")
	return err
}

func (r CLIRunner) Up(ctx context.Context, project, workDir, composeFile string, env []string, waitSec int, log func(string)) error {
	args := []string{"up", "-d", "--remove-orphans"}
	if waitSec > 0 {
		args = append(args, "--wait", "--wait-timeout", fmt.Sprintf("%d", waitSec))
	}
	_, err := r.runLogged(ctx, project, workDir, composeFile, env, log, args...)
	return err
}

func (r CLIRunner) Start(ctx context.Context, project, workDir, composeFile string) error {
	_, err := r.run(ctx, project, workDir, composeFile, nil, "start")
	return err
}

func (r CLIRunner) Stop(ctx context.Context, project, workDir, composeFile string) error {
	_, err := r.run(ctx, project, workDir, composeFile, nil, "stop")
	return err
}

func (r CLIRunner) Restart(ctx context.Context, project, workDir, composeFile string) error {
	_, err := r.run(ctx, project, workDir, composeFile, nil, "restart")
	return err
}

func (r CLIRunner) Down(ctx context.Context, project, workDir, composeFile string, volumes bool) error {
	args := []string{"down", "--remove-orphans"}
	if volumes {
		args = append(args, "-v")
	}
	_, err := r.run(ctx, project, workDir, composeFile, nil, args...)
	return err
}

func (r CLIRunner) PS(ctx context.Context, project, workDir, composeFile string) ([]ContainerStatus, error) {
	out, err := r.run(ctx, project, workDir, composeFile, nil, "ps", "--format", "json")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var statuses []ContainerStatus
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var raw struct {
			Service string `json:"Service"`
			State   string `json:"State"`
			Name    string `json:"Name"`
			Health  string `json:"Health"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		st := ContainerStatus{
			Service: raw.Service,
			State:   raw.State,
			Name:    raw.Name,
			Health:  raw.Health,
		}
		statuses = append(statuses, st)
	}
	return statuses, nil
}

func (r CLIRunner) run(ctx context.Context, project, workDir, composeFile string, env []string, args ...string) ([]byte, error) {
	return r.runLogged(ctx, project, workDir, composeFile, env, nil, args...)
}

func (r CLIRunner) runLogged(ctx context.Context, project, workDir, composeFile string, env []string, log func(string), args ...string) ([]byte, error) {
	if strings.TrimSpace(workDir) == "" {
		return nil, fmt.Errorf("work dir required")
	}
	if strings.TrimSpace(composeFile) == "" {
		return nil, fmt.Errorf("compose file required")
	}
	full := append([]string{"compose", "-p", project, "-f", composeFile}, args...)
	cmd := exec.CommandContext(ctx, "docker", full...)
	cmd.Dir = workDir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		out := strings.TrimSpace(buf.String())
		if log != nil && out != "" {
			for _, line := range strings.Split(out, "\n") {
				log(line)
			}
		}
		if out != "" {
			return nil, fmt.Errorf("%s: %s", err.Error(), out)
		}
		return nil, err
	}
	out := buf.Bytes()
	if log != nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.TrimSpace(line) != "" {
				log(line)
			}
		}
	}
	return out, nil
}

// PrepareReleaseDir writes compose.input.yaml and returns paths.
func PrepareReleaseDir(stackDir, deploymentID, yaml string) (releaseDir, inputPath string, err error) {
	releaseDir = filepath.Join(stackDir, "releases", deploymentID)
	if err := os.MkdirAll(releaseDir, 0o750); err != nil {
		return "", "", err
	}
	inputPath = filepath.Join(releaseDir, "compose.input.yaml")
	if err := os.WriteFile(inputPath, []byte(yaml), 0o600); err != nil {
		return "", "", err
	}
	return releaseDir, inputPath, nil
}

// WriteDockyardJSON writes the generated compose file.
func WriteDockyardJSON(releaseDir string, doc map[string]any) (string, error) {
	path := filepath.Join(releaseDir, "compose.dockyard.json")
	raw, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// WriteEnvFile writes .env for compose with mode 0600.
func WriteEnvFile(releaseDir string, lines []string) (string, error) {
	path := filepath.Join(releaseDir, ".env")
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// NormalizePipeline validates, mutates, and re-renders compose config.
func NormalizePipeline(ctx context.Context, runner Runner, cfg Config, in MutateInput, workDir, composeFile string, env []string, allowUnsafe bool) (map[string]any, []Issue, error) {
	doc, err := runner.ConfigJSON(ctx, cfg.ProjectName, workDir, composeFile, env)
	if err != nil {
		return nil, nil, err
	}
	opts := PolicyOptions{
		AllowUnsafe:           allowUnsafe,
		StackDataDir:          in.StackDataDir,
		PublishedPortsAllowed: in.AllowPublished,
	}
	issues := CheckPolicy(doc, opts)
	if !allowUnsafe && HasBlockingErrors(issues) {
		return doc, issues, fmt.Errorf("compose policy check failed")
	}
	doc, mutateIssues := Mutate(doc, in)
	issues = append(issues, mutateIssues...)
	if !allowUnsafe && HasBlockingErrors(issues) {
		return doc, issues, fmt.Errorf("compose policy check failed")
	}
	return doc, issues, nil
}
