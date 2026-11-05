package notify

import "time"

// Clock provides time for alert-state tests.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// AlertDecision is the outcome of stateful dedup logic.
type AlertDecision struct {
	Notify       bool
	Occurrences  int64
	UpdateState  bool
	ResolvedDrop bool
}

// DecideStateful applies firing/resolved rules with reminder and escalation.
func DecideStateful(
	prev *AlertStateRow,
	ev Event,
	now time.Time,
	reminder time.Duration,
) AlertDecision {
	if reminder <= 0 {
		reminder = defaultReminderInterval
	}
	dec := AlertDecision{Notify: true, Occurrences: 1, UpdateState: true}

	prevState := ""
	prevLevel := Level("")
	var lastNotified int64
	var occurrences int64 = 1
	if prev != nil {
		prevState = prev.State
		prevLevel = Level(prev.Level)
		lastNotified = prev.LastNotifiedAt
		occurrences = prev.Occurrences + 1
	}

	switch {
	case ev.State == StateFiring && (prev == nil || prevState == string(StateResolved)):
		dec.Occurrences = 1
		return dec
	case ev.State == StateFiring && prevState == string(StateFiring):
		dec.Occurrences = occurrences
		if levelRank(ev.Level) > levelRank(prevLevel) {
			dec.Notify = true
			return dec
		}
		if now.Sub(time.Unix(lastNotified, 0)) >= reminder {
			dec.Notify = true
			return dec
		}
		dec.Notify = false
		dec.UpdateState = true
		return dec
	case ev.State == StateResolved && prevState == string(StateFiring):
		dec.Occurrences = occurrences
		return dec
	case ev.State == StateResolved && prevState != string(StateFiring):
		dec.Notify = false
		dec.ResolvedDrop = true
		dec.UpdateState = false
		return dec
	default:
		dec.Notify = false
		dec.UpdateState = false
		return dec
	}
}

// AlertStateRow mirrors DB alert_state for pure logic.
type AlertStateRow struct {
	State          string
	Level          string
	LastNotifiedAt int64
	Occurrences    int64
}
