package appconfig

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	BuilderDockerfile = "dockerfile"
	BuilderRailpack   = "railpack"
)

// EffectiveBuilder picks dockerfile vs railpack after clone (briefing §11.1).
func (c AppConfig) EffectiveBuilder(contextDir string) string {
	explicit := strings.ToLower(strings.TrimSpace(c.Builder))
	if explicit == BuilderDockerfile || explicit == BuilderRailpack {
		return explicit
	}
	df := filepath.Join(contextDir, c.DockerfilePath())
	if info, err := os.Stat(df); err == nil && !info.IsDir() {
		return BuilderDockerfile
	}
	return BuilderRailpack
}

func (c AppConfig) validateGitBuilder() error {
	builder := strings.ToLower(strings.TrimSpace(c.Builder))
	if builder == "" {
		return nil
	}
	if builder != BuilderDockerfile && builder != BuilderRailpack {
		return errUnsupportedBuilder(c.Builder)
	}
	return nil
}

func errUnsupportedBuilder(b string) error {
	return &BuilderError{Builder: b}
}

// BuilderError indicates an unknown builder string.
type BuilderError struct {
	Builder string
}

func (e *BuilderError) Error() string {
	return "builder " + e.Builder + " not supported"
}

// NeedsGitBuild reports whether deployment should clone and build.
func (c AppConfig) NeedsGitBuild() bool {
	t := strings.ToLower(strings.TrimSpace(c.Source.Type))
	return t == "git" || t == "github"
}

// CloneRemote returns URL and branch for shallow clone.
func (s Source) CloneRemote() (url, branch string, err error) {
	switch strings.ToLower(strings.TrimSpace(s.Type)) {
	case "git":
		url = strings.TrimSpace(s.URL)
		if url == "" {
			return "", "", errGitURLRequired
		}
		return url, s.BranchOrDefault(), nil
	case "github":
		repo := strings.TrimSpace(s.Repo)
		if repo == "" {
			return "", "", errGitHubRepoRequired
		}
		return "https://github.com/" + strings.Trim(repo, "/") + ".git", s.BranchOrDefault(), nil
	default:
		return "", "", errNotGitSource
	}
}

// BranchOrDefault returns branch or main.
func (s Source) BranchOrDefault() string {
	b := strings.TrimSpace(s.Branch)
	if b == "" {
		return "main"
	}
	return b
}

// GitHubRepoFullName returns owner/name for webhooks.
func (s Source) GitHubRepoFullName() string {
	return strings.TrimSpace(s.Repo)
}

var (
	errGitURLRequired     = errString("git url required")
	errGitHubRepoRequired = errString("github repo required")
	errNotGitSource       = errString("not a git-based source")
)

type errString string

func (e errString) Error() string { return string(e) }
