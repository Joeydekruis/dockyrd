package hostops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

// TestEcho is a minimal real action for unit tests (marker file under DataDir).
type TestEcho struct {
	Cfg            config.Config
	RequireLockout bool
	markerName     string
}

func NewTestEcho(cfg config.Config) *TestEcho {
	return &TestEcho{Cfg: cfg, markerName: "hostops-test-echo.marker"}
}

func (t *TestEcho) WithLockoutConfirm() *TestEcho {
	t.RequireLockout = true
	return t
}

func (t *TestEcho) Kind() string { return KindTestEcho }

func (t *TestEcho) markerPath() string {
	return filepath.Join(t.Cfg.DataDir, t.markerName)
}

func (t *TestEcho) Plan(_ context.Context) (Plan, error) {
	return Plan{
		"message": "test echo preview",
		"path":    t.markerPath(),
	}, nil
}

func (t *TestEcho) Apply(_ context.Context) (RevertPayload, error) {
	path := t.markerPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, err
	}
	prev := ""
	if data, err := os.ReadFile(path); err == nil {
		prev = string(data)
	}
	if err := os.WriteFile(path, []byte("applied"), 0o600); err != nil {
		return nil, err
	}
	b, err := json.Marshal(map[string]string{"path": path, "previous": prev})
	if err != nil {
		return nil, err
	}
	return RevertPayload(b), nil
}

func (t *TestEcho) Verify(_ context.Context) error {
	data, err := os.ReadFile(t.markerPath())
	if err != nil {
		return err
	}
	if string(data) != "applied" {
		return fmt.Errorf("marker content mismatch")
	}
	return nil
}

func (t *TestEcho) Revert(_ context.Context, payload RevertPayload) error {
	var meta struct {
		Path     string `json:"path"`
		Previous string `json:"previous"`
	}
	if err := json.Unmarshal(payload, &meta); err != nil {
		return err
	}
	if meta.Previous == "" {
		return os.Remove(meta.Path)
	}
	return os.WriteFile(meta.Path, []byte(meta.Previous), 0o600)
}

func (t *TestEcho) RequiresLockoutConfirm() bool { return t.RequireLockout }
