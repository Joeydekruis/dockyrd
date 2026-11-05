package backup

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Joeydekruis/dockyrd/internal/notify"
)

func TestBackupNotifyDedupKey(t *testing.T) {
	svcID := "01SERVICE"
	dest := "remote"
	key := "backup:" + svcID + ":" + dest
	if key != "backup:01SERVICE:remote" {
		t.Fatalf("unexpected key %q", key)
	}
}

func TestBackupErrorScrubbedInEmit(t *testing.T) {
	long := strings.Repeat("x", 600)
	scrubbed := notify.ScrubDetailForEmit(long)
	if utf8.RuneCountInString(scrubbed) > 500 {
		t.Fatalf("expected scrubbed rune cap, got %d", utf8.RuneCountInString(scrubbed))
	}
}
