package staticsite

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
)

// Config is services.config for kind=static_site.
type Config struct {
	Source      appconfig.Source `json:"source"`
	AutoDeploy  bool             `json:"autoDeploy"`
	BuildCmd    string           `json:"buildCmd"`
	OutputDir   string           `json:"outputDir"`
	NodeImage   string           `json:"nodeImage"`
	SpaFallback bool             `json:"spaFallback"`
}

// Parse decodes static site service config.
func Parse(raw json.RawMessage) (Config, error) {
	var cfg Config
	if len(raw) == 0 {
		return Config{}, errors.New("static site config required")
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.NodeImage == "" {
		cfg.NodeImage = "node:22-alpine"
	}
	if cfg.BuildCmd == "" {
		cfg.BuildCmd = "npm ci && npm run build"
	}
	t := strings.ToLower(strings.TrimSpace(cfg.Source.Type))
	if t != "git" && t != "github" {
		return Config{}, errors.New("static site requires git or github source")
	}
	if _, _, err := cfg.Source.CloneRemote(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
