package preview

import (
	"encoding/json"
	"testing"
)

func TestEnabledRaw(t *testing.T) {
	raw := map[string]any{"preview": Config{Enabled: true}}
	if !EnabledRaw(raw) {
		t.Fatal("expected enabled")
	}
	if EnabledRaw(map[string]any{"preview": Config{Enabled: false}}) {
		t.Fatal("expected disabled")
	}
}

func TestMergeConfig(t *testing.T) {
	out, err := MergeConfig(json.RawMessage(`{"port":80}`), Config{Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if !Enabled(out) {
		t.Fatal("merge should enable preview")
	}
}
