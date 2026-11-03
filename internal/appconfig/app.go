package appconfig

import (
	"encoding/json"
	"errors"
	"strings"
)

// AppConfig is the JSON stored on services.config for kind=app.
type AppConfig struct {
	Source     Source `json:"source"`
	Port       int    `json:"port"`
	HealthPath string `json:"healthPath"`
	MemLimitMB int    `json:"memLimitMB"`
	StopGrace  int    `json:"stopGrace"`
}

// Source describes how to obtain a runnable image.
type Source struct {
	Type string `json:"type"`
	Ref  string `json:"ref"`
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
	return cfg, nil
}

// ImageRef returns the docker image reference for image-sourced apps.
func (c AppConfig) ImageRef() (string, error) {
	if strings.ToLower(c.Source.Type) != "image" {
		return "", errors.New("only image source supported in M1")
	}
	ref := strings.TrimSpace(c.Source.Ref)
	if ref == "" {
		return "", errors.New("image ref required")
	}
	return ref, nil
}
