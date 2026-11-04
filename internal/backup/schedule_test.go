package backup

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNextRunAfter(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if got := NextRunAfter(base, "hourly"); !got.Equal(base.Add(time.Hour)) {
		t.Fatalf("hourly: %v", got)
	}
	if got := NextRunAfter(base, "daily"); !got.Equal(base.Add(24 * time.Hour)) {
		t.Fatalf("daily: %v", got)
	}
	if got := NextRunAfter(base, "weekly"); !got.Equal(base.Add(7 * 24 * time.Hour)) {
		t.Fatalf("weekly: %v", got)
	}
}

func TestJobPayloadRoundTrip(t *testing.T) {
	in := JobPayload{ServiceID: "01ABC", PolicyID: "01DEF"}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out JobPayload
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("got %+v", out)
	}
}
