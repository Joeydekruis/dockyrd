package hostops

import "testing"

func TestFSMValidTransitions(t *testing.T) {
	valid := []struct{ from, to State }{
		{StatePlanned, StateApplied},
		{StatePlanned, StateCancelled},
		{StateApplied, StateVerified},
		{StateApplied, StateFailed},
		{StateApplied, StateReverted},
		{StateVerified, StateAwaitingConfirm},
		{StateVerified, StateConfirmed},
		{StateVerified, StateReverted},
		{StateAwaitingConfirm, StateConfirmed},
		{StateAwaitingConfirm, StateReverted},
		{StateFailed, StateReverted},
	}
	for _, tc := range valid {
		if err := ValidateTransition(tc.from, tc.to); err != nil {
			t.Errorf("%s -> %s: unexpected error: %v", tc.from, tc.to, err)
		}
	}
}

func TestFSMInvalidTransitions(t *testing.T) {
	invalid := []struct{ from, to State }{
		{StatePlanned, StateVerified},
		{StateReverted, StateApplied},
		{StateCancelled, StateApplied},
		{StatePlanned, StateConfirmed},
	}
	for _, tc := range invalid {
		if err := ValidateTransition(tc.from, tc.to); err == nil {
			t.Errorf("%s -> %s: expected error", tc.from, tc.to)
		}
	}
}

func TestAuditActionForState(t *testing.T) {
	if AuditActionForTransition(StatePlanned, StateApplied) != "host_action.applied" {
		t.Fatal("bad audit name for applied")
	}
}
