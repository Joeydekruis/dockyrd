package backup

import "time"

// NextRunAfter returns the next scheduled run from a completed backup time.
func NextRunAfter(from time.Time, interval string) time.Time {
	switch interval {
	case "hourly":
		return from.Add(time.Hour)
	case "weekly":
		return from.Add(7 * 24 * time.Hour)
	default:
		return from.Add(24 * time.Hour)
	}
}
