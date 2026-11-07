package securitypolicy

// EffectiveProfile resolves inherit → instance default.
func EffectiveProfile(inst Instance, site Site) Profile {
	switch site.Profile {
	case ProfileInherit:
		return inst.DefaultProfile
	case ProfileBaseline, ProfileStrict, ProfileCustom:
		return site.Profile
	default:
		return inst.DefaultProfile
	}
}

// profileRank returns higher = stricter.
func profileRank(p Profile) int {
	switch p {
	case ProfileStrict:
		return 3
	case ProfileCustom:
		return 2
	case ProfileBaseline:
		return 1
	default:
		return 0
	}
}

// CheckFloor rejects weaker site settings when instance lock is on (unless lock_exempt).
func CheckFloor(inst Instance, site Site, in FloorInput, lockExempt bool) *FloorError {
	if !inst.Lock || lockExempt || site.LockExempt {
		return nil
	}
	floor := inst.DefaultProfile
	candidate := in.Profile
	if candidate == ProfileInherit {
		candidate = inst.DefaultProfile
	}
	if profileRank(candidate) < profileRank(floor) {
		return &FloorError{
			Field:   "profile",
			Message: "profile cannot be weaker than instance default while policy lock is enabled",
		}
	}
	return nil
}
