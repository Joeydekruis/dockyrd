package compose

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	SourceInline = "inline"
	SourceGit    = "git"
	defaultWait  = 300
)

type Config struct {
	Source                ComposeSource `json:"source"`
	ProjectName           string        `json:"project_name,omitempty"`
	PublishedPortsAllowed []string      `json:"published_ports_allowed,omitempty"`
	WaitTimeoutSec        int           `json:"wait_timeout_sec,omitempty"`
	LastServices          []string      `json:"last_services,omitempty"`
}

type ComposeSource struct {
	Type string `json:"type"`
	YAML string `json:"yaml,omitempty"`
}

type Exposure struct {
	Service string `json:"service"`
	Port    int    `json:"port"`
}

func Parse(raw json.RawMessage) (Config, error) {
	if len(raw) == 0 {
		return Config{}, errors.New("config required")
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, err
	}
	cfg.Source.Type = strings.TrimSpace(strings.ToLower(cfg.Source.Type))
	if cfg.Source.Type != SourceInline {
		return Config{}, fmt.Errorf("unsupported source.type %q", cfg.Source.Type)
	}
	if strings.TrimSpace(cfg.Source.YAML) == "" {
		return Config{}, errors.New("source.yaml required")
	}
	if cfg.WaitTimeoutSec <= 0 {
		cfg.WaitTimeoutSec = defaultWait
	}
	if strings.TrimSpace(cfg.ProjectName) == "" {
		return Config{}, errors.New("project_name missing")
	}
	return cfg, nil
}

func NewInlineConfigJSON(serviceID, yaml string) (json.RawMessage, error) {
	yaml = strings.TrimSpace(yaml)
	if yaml == "" {
		return nil, errors.New("yaml required")
	}
	return json.Marshal(Config{
		Source:         ComposeSource{Type: SourceInline, YAML: yaml},
		ProjectName:    ProjectName(serviceID),
		WaitTimeoutSec: defaultWait,
	})
}

func ProjectName(serviceID string) string {
	short := serviceID
	if len(short) > 8 {
		short = short[:8]
	}
	return "dyc-" + strings.ToLower(short)
}

func StackNetworkName(project string) string { return project + "_default" }

func ServiceAlias(serviceID, name string) string {
	return ProjectName(serviceID) + "-" + strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), "_", "-"))
}

func UpstreamDial(serviceID, composeService string, port int) string {
	return fmt.Sprintf("%s:%d", ServiceAlias(serviceID, composeService), port)
}

func WithTimeout(ctx context.Context, waitSec int) (context.Context, context.CancelFunc) {
	if waitSec <= 0 {
		waitSec = defaultWait
	}
	return context.WithTimeout(ctx, time.Duration(waitSec)*time.Second)
}
