package setup

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
)

func TestInitCreatesTokenAndMasterKey(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:   filepath.Join(dir, "data"),
		ConfigDir: filepath.Join(dir, "etc"),
		Listen:    ":8000",
	}

	res, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if res.AlreadyInitialized {
		t.Fatal("expected first init")
	}
	if res.Token == "" {
		t.Fatal("expected setup token")
	}
	if !res.MasterKeyCreated {
		t.Fatal("expected master key created")
	}

	res2, err := Init(cfg)
	if err != nil {
		t.Fatalf("Init again: %v", err)
	}
	if res2.Token != "" {
		t.Fatal("second init should not re-print token")
	}
}

func TestIssueSetupToken(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:   filepath.Join(dir, "data"),
		ConfigDir: filepath.Join(dir, "etc"),
		Listen:    ":8000",
	}
	if _, err := Init(cfg); err != nil {
		t.Fatal(err)
	}

	plain, exp, err := IssueSetupToken(cfg)
	if err != nil {
		t.Fatalf("IssueSetupToken: %v", err)
	}
	if plain == "" || time.Until(exp) < time.Minute {
		t.Fatalf("token=%q exp=%v", plain, exp)
	}
}

func TestGenerateTokenFormat(t *testing.T) {
	plain, setting, err := GenerateToken(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(plain) < 10 || setting.TokenHash == "" {
		t.Fatalf("plain=%q setting=%+v", plain, setting)
	}
}

func TestIsCompleteFalse(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:   filepath.Join(dir, "data"),
		ConfigDir: filepath.Join(dir, "etc"),
	}
	if _, err := Init(cfg); err != nil {
		t.Fatal(err)
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ok, err := IsComplete(context.Background(), q)
	if err != nil || ok {
		t.Fatalf("complete=%v err=%v", ok, err)
	}
}
