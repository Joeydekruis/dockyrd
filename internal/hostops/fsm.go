package hostops

import (
	"errors"
	"fmt"
)

// State is a host_actions lifecycle value (briefing Bijlage B).
type State string

const (
	StatePlanned         State = "planned"
	StateApplied         State = "applied"
	StateVerified        State = "verified"
	StateAwaitingConfirm State = "awaiting_confirm"
	StateConfirmed       State = "confirmed"
	StateReverted        State = "reverted"
	StateFailed          State = "failed"
	StateCancelled       State = "cancelled"
)

var ErrInvalidTransition = errors.New("invalid host action state transition")

// ValidateTransition returns nil when from -> to is allowed.
func ValidateTransition(from, to State) error {
	if from == to {
		return nil
	}
	allowed := transitionTable[from]
	if allowed == nil || !allowed[to] {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, from, to)
	}
	return nil
}

var transitionTable = map[State]map[State]bool{
	StatePlanned: {
		StateApplied:   true,
		StateCancelled: true,
	},
	StateApplied: {
		StateVerified: true,
		StateFailed:   true,
		StateReverted: true,
	},
	StateVerified: {
		StateAwaitingConfirm: true,
		StateConfirmed:       true,
		StateReverted:        true,
	},
	StateAwaitingConfirm: {
		StateConfirmed: true,
		StateReverted:  true,
	},
	StateConfirmed: {
		StateReverted: true,
	},
	StateFailed: {
		StateReverted: true,
	},
}

// Revertable reports whether manual or timer revert is meaningful from state.
func Revertable(state State) bool {
	switch state {
	case StateApplied, StateVerified, StateAwaitingConfirm, StateConfirmed, StateFailed:
		return true
	default:
		return false
	}
}

// AuditActionForTransition returns the audit_log action for entering `to`.
func AuditActionForTransition(_, to State) string {
	return "host_action." + string(to)
}
