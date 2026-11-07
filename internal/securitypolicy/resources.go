package securitypolicy

// EffectiveMemMB returns deploy memory limit (MiB).
func EffectiveMemMB(inst Instance, site Site) int {
	if site.MemMB != nil && *site.MemMB > 0 {
		return *site.MemMB
	}
	if inst.DefaultMemMB > 0 {
		return inst.DefaultMemMB
	}
	return 512
}

// EffectiveCPU returns vCPU limit.
func EffectiveCPU(inst Instance, site Site) float64 {
	if site.CPU != nil && *site.CPU > 0 {
		return *site.CPU
	}
	if inst.DefaultCPU > 0 {
		return inst.DefaultCPU
	}
	return 1.0
}

// EffectivePIDs returns pids limit.
func EffectivePIDs(inst Instance, site Site) int {
	if site.PIDs != nil && *site.PIDs > 0 {
		return *site.PIDs
	}
	if inst.DefaultPIDs > 0 {
		return inst.DefaultPIDs
	}
	return 512
}

// ResourcesChanged reports whether mem/cpu/pids differ from stored site row.
func ResourcesChanged(before Site, after Site) bool {
	return !ptrIntEq(before.MemMB, after.MemMB) ||
		!ptrFloatEq(before.CPU, after.CPU) ||
		!ptrIntEq(before.PIDs, after.PIDs)
}

func ptrIntEq(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrFloatEq(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
