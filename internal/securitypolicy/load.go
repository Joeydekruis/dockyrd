package securitypolicy

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// LoadInstance reads the singleton policy row.
func LoadInstance(ctx context.Context, q *sqlc.Queries) (Instance, error) {
	row, err := q.GetSecurityPolicy(ctx)
	if err != nil {
		return Instance{}, err
	}
	return InstanceFromRow(row)
}

// LoadSite reads or defaults service_security for a service.
func LoadSite(ctx context.Context, q *sqlc.Queries, serviceID string) (Site, error) {
	row, err := q.GetServiceSecurity(ctx, serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DefaultSite(serviceID), nil
		}
		return Site{}, err
	}
	return SiteFromRow(row)
}

// SnapshotForService resolves proxy snapshot for a service id.
func SnapshotForService(ctx context.Context, q *sqlc.Queries, serviceID string) (RouteSnapshot, error) {
	inst, err := LoadInstance(ctx, q)
	if err != nil {
		return RouteSnapshot{}, err
	}
	site, err := LoadSite(ctx, q, serviceID)
	if err != nil {
		return RouteSnapshot{}, err
	}
	return SnapshotForRoute(inst, site), nil
}
