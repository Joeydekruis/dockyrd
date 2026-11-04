package serviceops

import (
	"context"

	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// StopManagedDatabaseContainer stops and removes the running container for a database service.
func StopManagedDatabaseContainer(ctx context.Context, dc *docker.Client, serviceID string) error {
	if dc == nil {
		return nil
	}
	managed, err := dc.ListManaged(ctx)
	if err != nil {
		return err
	}
	for _, c := range managed {
		if c.Identity.ServiceID != serviceID {
			continue
		}
		if c.Identity.DeploymentID != "" {
			continue
		}
		if c.Identity.Kind != docker.KindDatabase && c.Identity.Kind != docker.KindRedis {
			continue
		}
		return dc.StopAndRemove(ctx, c.ID, 30)
	}
	return nil
}

// TeardownDatabase stops the container and optionally removes its data volume.
func TeardownDatabase(ctx context.Context, dc *docker.Client, serviceID, serviceSlug string, removeVolume bool) error {
	if err := StopManagedDatabaseContainer(ctx, dc, serviceID); err != nil {
		return err
	}
	if removeVolume && dc != nil {
		return dc.RemoveVolume(ctx, docker.VolumeName(serviceSlug))
	}
	return nil
}
