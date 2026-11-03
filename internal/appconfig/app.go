package appconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
)

// AppConfig is the JSON stored on services.config for kind=app.
type AppConfig struct {
	Source     Source `json:"source"`
	Builder    string `json:"builder"`
	Dockerfile string `json:"dockerfile"`
	AutoDeploy bool   `json:"autoDeploy"`
	Port       int    `json:"port"`
	HealthPath string `json:"healthPath"`
	MemLimitMB int    `json:"memLimitMB"`
	StopGrace  int    `json:"stopGrace"`
}

// Source describes how to obtain a runnable image.
type Source struct {
	Type    string `json:"type"`
	Ref     string `json:"ref"`
	URL     string `json:"url"`
	Repo    string `json:"repo"`
	Branch  string `json:"branch"`
	RootDir string `json:"rootDir"`
}

// Parse decodes and validates an app service config blob.
func Parse(raw json.RawMessage) (AppConfig, error) {
	var cfg AppConfig
	if len(raw) == 0 {
		raw = json.RawMessage(`{"source":{"type":"image","ref":"nginx:alpine"},"port":80}`)
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return AppConfig{}, err
	}
	if cfg.Port <= 0 {
		cfg.Port = 80
	}
	if cfg.MemLimitMB <= 0 {
		cfg.MemLimitMB = 512
	}
	if cfg.StopGrace <= 0 {
		cfg.StopGrace = 30
	}
	if cfg.Dockerfile == "" {
		cfg.Dockerfile = "Dockerfile"
	}
	return cfg, nil
}

// ImageRef returns the docker image reference for image-sourced apps.
func (c AppConfig) ImageRef() (string, error) {
	ref, _, err := c.DeploymentImageRef("", "")
	return ref, err
}

// DeploymentImageRef resolves the image to run for a deployment.
func (c AppConfig) DeploymentImageRef(serviceSlug, deployShort string) (ref string, needsBuild bool, err error) {
	switch strings.ToLower(strings.TrimSpace(c.Source.Type)) {
	case "image", "":
		if strings.ToLower(c.Source.Type) == "" {
			c.Source.Type = "image"
		}
		ref = strings.TrimSpace(c.Source.Ref)
		if ref == "" {
			return "", false, errors.New("image ref required")
		}
		return ref, false, nil
	case "git", "github":
		if err := c.validateGitBuilder(); err != nil {
			return "", false, err
		}
		serviceSlug = strings.TrimSpace(serviceSlug)
		deployShort = strings.TrimSpace(deployShort)
		if serviceSlug == "" || deployShort == "" {
			return "", false, errors.New("service slug and deploy id required for git builds")
		}
		tag := fmt.Sprintf("dockyard/%s:%s", serviceSlug, deployShort)
		return tag, true, nil
	default:
		return "", false, fmt.Errorf("unsupported source type %q", c.Source.Type)
	}
}

// GitCloneURL returns the remote URL for shallow clone.
func (c AppConfig) GitCloneURL() (string, error) {
	url, _, err := c.Source.CloneRemote()
	return url, err
}

// GitBranch returns the branch to clone (default main).
func (c AppConfig) GitBranch() string {
	return c.Source.BranchOrDefault()
}

// BuildContextDir returns the subdirectory within the clone to use as docker context.
func (c AppConfig) BuildContextDir(cloneRoot string) string {
	root := strings.TrimSpace(c.Source.RootDir)
	if root == "" || root == "." {
		return cloneRoot
	}
	clean := path.Clean(root)
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return cloneRoot
	}
	return path.Join(cloneRoot, clean)
}

// DockerfilePath returns the Dockerfile path relative to build context.
func (c AppConfig) DockerfilePath() string {
	if strings.TrimSpace(c.Dockerfile) == "" {
		return "Dockerfile"
	}
	return path.Clean(c.Dockerfile)
}
