package deploy

import (
	"context"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// appAttachEdge reports whether an app container should join dy-edge (public routing).
func (rt *Runtime) appAttachEdge(ctx context.Context, serviceID string) (bool, error) {
	domains, err := rt.Q.ListDomainsForService(ctx, serviceID)
	if err != nil {
		return false, err
	}
	return len(domains) > 0, nil
}

// AppAttachEdge is exported for tests and callers outside Runtime methods.
func AppAttachEdge(ctx context.Context, q *sqlc.Queries, serviceID string) (bool, error) {
	domains, err := q.ListDomainsForService(ctx, serviceID)
	if err != nil {
		return false, err
	}
	return len(domains) > 0, nil
}
