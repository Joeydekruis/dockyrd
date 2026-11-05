package links

import (
	"errors"
	"fmt"
)

var (
	ErrSameService   = errors.New("cannot link service to itself")
	ErrInvalidFrom   = errors.New("consumer must be an application")
	ErrInvalidTo     = errors.New("invalid link target kind")
	ErrStaticTarget  = errors.New("static sites cannot be link targets")
	ErrStaticFrom    = errors.New("static sites cannot create links")
)

// ValidConsumerKinds are service kinds allowed as link consumers (from).
func ValidConsumerKinds() map[string]bool {
	return map[string]bool{"app": true}
}

// ValidProviderKinds are service kinds allowed as link targets (to).
func ValidProviderKinds() map[string]bool {
	return map[string]bool{
		"app":      true,
		"postgres": true,
		"redis":    true,
		"bucket":   true,
	}
}

// ValidatePair checks directed link kinds per briefing D2.
func ValidatePair(fromKind, toKind string) error {
	if fromKind == "static_site" {
		return ErrStaticFrom
	}
	if !ValidConsumerKinds()[fromKind] {
		return fmt.Errorf("%w: %q", ErrInvalidFrom, fromKind)
	}
	if toKind == "static_site" {
		return ErrStaticTarget
	}
	if !ValidProviderKinds()[toKind] {
		return fmt.Errorf("%w: %q", ErrInvalidTo, toKind)
	}
	return nil
}
