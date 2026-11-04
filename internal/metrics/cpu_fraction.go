package metrics

// NormalizeHostCPUFraction converts stored CPU to fraction 0–1 for API responses.
// Legacy collectors wrote 0–100; current writers use 0–1.
func NormalizeHostCPUFraction(cpu float64) float64 {
	if cpu > 1 {
		return cpu / 100
	}
	return cpu
}
