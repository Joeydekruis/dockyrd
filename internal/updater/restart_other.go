//go:build !linux

package updater

// ScheduleServiceRestart is a no-op off Linux (dev hosts).
func ScheduleServiceRestart() {}
