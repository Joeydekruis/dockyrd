package watch

import (
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/backup"
)

func TestBackupOverdueGrace(t *testing.T) {
	last := time.Now().Add(-48 * time.Hour)
	nextDue := backup.NextRunAfterLegacy(last, "daily")
	grace := nextDue.Add(30 * time.Minute)
	if !time.Now().After(grace) {
		t.Fatal("expected overdue after 2 days without daily backup")
	}
}

func TestFormatLastNever(t *testing.T) {
	if formatLast(time.Time{}) != "never" {
		t.Fatal("zero time should format as never")
	}
}

func TestBackupScheduleOKWithinGrace(t *testing.T) {
	last := time.Now().Add(-2 * time.Hour)
	nextDue := backup.NextRunAfterLegacy(last, "daily")
	grace := nextDue.Add(30 * time.Minute)
	if time.Now().After(grace) {
		t.Fatal("recent backup should not be overdue")
	}
}
