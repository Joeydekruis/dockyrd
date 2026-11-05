package backup

import "time"

// ScheduleParams controls when the next backup run is due.
type ScheduleParams struct {
	Interval      string
	RunAtMinute   int64 // minutes from midnight UTC (0–1439)
	RunWeekday    int64 // 0=Sunday, used when Interval is weekly
}

// NextRunAfter returns the next scheduled run strictly after from.
func NextRunAfter(from time.Time, p ScheduleParams) time.Time {
	from = from.UTC()
	switch p.Interval {
	case "hourly":
		return from.Add(time.Hour)
	case "weekly":
		return nextWeekly(from, p.RunAtMinute, p.RunWeekday)
	default:
		return nextDaily(from, p.RunAtMinute)
	}
}

func nextDaily(from time.Time, runAtMinute int64) time.Time {
	if runAtMinute < 0 {
		runAtMinute = 0
	}
	if runAtMinute > 1439 {
		runAtMinute = 1439
	}
	hour := runAtMinute / 60
	min := runAtMinute % 60
	candidate := time.Date(from.Year(), from.Month(), from.Day(), int(hour), int(min), 0, 0, time.UTC)
	if !candidate.After(from) {
		candidate = candidate.Add(24 * time.Hour)
	}
	return candidate
}

func nextWeekly(from time.Time, runAtMinute, runWeekday int64) time.Time {
	if runWeekday < 0 {
		runWeekday = 0
	}
	if runWeekday > 6 {
		runWeekday = 6
	}
	want := time.Weekday(runWeekday)
	candidate := nextDaily(from, runAtMinute)
	for i := 0; i < 8; i++ {
		if candidate.Weekday() == want && candidate.After(from) {
			return candidate
		}
		candidate = candidate.Add(24 * time.Hour)
	}
	return candidate
}

// ScheduleFromInterval builds params from policy columns.
func ScheduleFromInterval(interval string, runAtMinute, runWeekday int64) ScheduleParams {
	return ScheduleParams{
		Interval:    interval,
		RunAtMinute: runAtMinute,
		RunWeekday:  runWeekday,
	}
}

// NextRunAfterLegacy wraps interval-only scheduling (defaults 03:00 UTC daily).
func NextRunAfterLegacy(from time.Time, interval string) time.Time {
	return NextRunAfter(from, ScheduleFromInterval(interval, 180, 0))
}
