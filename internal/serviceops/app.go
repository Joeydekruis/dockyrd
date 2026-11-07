package serviceops

import (
	"context"

	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// TeardownApplication stops and removes all managed app containers for a service.
func TeardownApplication(ctx context.Context, dc *docker.Client, serviceID string, graceSeconds int) error {
	if dc == nil {
		return nil
	}
	if graceSeconds <= 0 {
		graceSeconds = 30
	}
	managed, err := dc.ListManaged(ctx)
	if err != nil {
		return err
	}
	var firstErr error
	for _, c := range managed {
		if c.Identity.ServiceID != serviceID {
			continue
		}
		if c.Identity.Kind != docker.KindApp {
			continue
		}
		if err := dc.StopAndRemove(ctx, c.ID, graceSeconds); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
