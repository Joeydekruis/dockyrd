package backup

import (
	"testing"
	"time"
)

func TestNextRunAfter(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if got := NextRunAfter(base, ScheduleParams{Interval: "hourly"}); !got.Equal(base.Add(time.Hour)) {
		t.Fatalf("hourly: %v", got)
	}
	daily := NextRunAfter(base, ScheduleParams{Interval: "daily", RunAtMinute: 180})
	wantDaily := time.Date(2026, 1, 2, 3, 0, 0, 0, time.UTC)
	if !daily.Equal(wantDaily) {
		t.Fatalf("daily at 03:00: got %v want %v", daily, wantDaily)
	}
	weeklyBase := time.Date(2026, 1, 5, 10, 0, 0, 0, time.UTC) // Monday
	weekly := NextRunAfter(weeklyBase, ScheduleParams{
		Interval: "weekly", RunAtMinute: 180, RunWeekday: 0,
	})
	wantWeekly := time.Date(2026, 1, 11, 3, 0, 0, 0, time.UTC) // next Sunday
	if !weekly.Equal(wantWeekly) {
		t.Fatalf("weekly: got %v want %v", weekly, wantWeekly)
	}
}

func TestNextRunAfterSameDayDaily(t *testing.T) {
	base := time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)
	got := NextRunAfter(base, ScheduleParams{Interval: "daily", RunAtMinute: 180})
	want := time.Date(2026, 1, 1, 3, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
