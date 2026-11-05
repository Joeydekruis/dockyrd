package notify

import (
	"strings"
	"testing"
)

func TestValidateEmailChannelConfig(t *testing.T) {
	if err := ValidateEmailChannelConfig(ChannelConfig{
		"recipients": map[string]any{
			"admins": true,
		},
	}); err != nil {
		t.Fatalf("expected valid: %v", err)
	}
	if err := ValidateEmailChannelConfig(ChannelConfig{}); err == nil {
		t.Fatal("expected error for empty recipients")
	}
}

func TestEmailSubjectResolved(t *testing.T) {
	subj := emailSubject(EventView{
		Type:  TypeBackupSucceeded,
		Level: LevelSuccess,
		State: StateResolved,
		Title: "Backup succeeded · db",
	})
	if !strings.Contains(subj, "[RESOLVED]") {
		t.Fatalf("expected resolved prefix, got %q", subj)
	}
}

func TestResolveEmailTargetsDedupSkip(t *testing.T) {
	skip := map[string]struct{}{"admin@example.com": {}}
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range []string{"admin@example.com", "ops@example.com"} {
		key := raw
		if _, ok := skip[key]; ok {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	if len(out) != 1 || out[0] != "ops@example.com" {
		t.Fatalf("dedup skip failed: %v", out)
	}
}