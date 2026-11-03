package metrics

import "time"

// HostSample is one host metric row (1-minute resolution).
type HostSample struct {
	TS        int64
	CPU       float64
	MemUsed   int64
	MemTotal  int64
	DiskUsed  int64
	DiskTotal int64
	NetRx     int64
	NetTx     int64
	Load1     float64
}

// MinuteTS returns unix timestamp truncated to minute boundary.
func MinuteTS(t time.Time) int64 {
	return t.UTC().Truncate(time.Minute).Unix()
}
