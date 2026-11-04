package staticsite

import "testing"

func TestResolveBuildCommand_explicitEmpty(t *testing.T) {
	cfg := Config{InstallCmd: "", BuildScript: "", BuildCmd: ""}
	if got := ResolveBuildCommand(cfg, t.TempDir()); got != "" {
		t.Fatalf("got %q want empty", got)
	}
}

func TestResolveBuildCommand_sanitizeFalse(t *testing.T) {
	cfg := Config{BuildCmd: "false"}
	if got := ResolveBuildCommand(cfg, t.TempDir()); got != "" {
		t.Fatalf("got %q want empty", got)
	}
}

func TestResolveBuildCommand_splitFields(t *testing.T) {
	cfg := Config{InstallCmd: "npm ci", BuildScript: "npm run build"}
	if got := ResolveBuildCommand(cfg, t.TempDir()); got != "npm ci && npm run build" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveBuildCommand_installOnly(t *testing.T) {
	cfg := Config{InstallCmd: "npm ci", BuildScript: ""}
	if got := ResolveBuildCommand(cfg, t.TempDir()); got != "npm ci" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveBuildCommand_buildDisabled(t *testing.T) {
	disabled := false
	cfg := Config{BuildEnabled: &disabled, BuildCmd: "npm ci && npm run build"}
	if got := ResolveBuildCommand(cfg, t.TempDir()); got != "" {
		t.Fatalf("got %q want empty when build disabled", got)
	}
}
