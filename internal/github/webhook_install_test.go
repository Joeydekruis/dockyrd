package github

import (
	"testing"
)

func TestParseInstallationEvent(t *testing.T) {
	body := []byte(`{"action":"created","installation":{"id":12345}}`)
	ev, err := ParseInstallationEvent(body)
	if err != nil || ev.Installation.ID != 12345 || ev.Action != "created" {
		t.Fatalf("got %+v err=%v", ev, err)
	}
}

func TestParseInstallationEventMissingID(t *testing.T) {
	_, err := ParseInstallationEvent([]byte(`{"action":"created"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}
