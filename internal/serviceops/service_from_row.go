package serviceops

import "github.com/Joeydekruis/dockyrd/internal/db/sqlc"

// ServiceFromListRow maps ListServicesByProject rows to Service.
func ServiceFromListRow(row sqlc.ListServicesByProjectRow) sqlc.Service {
	return sqlc.Service{
		ID:               row.ID,
		EnvironmentID:    row.EnvironmentID,
		Kind:             row.Kind,
		Name:             row.Name,
		Slug:             row.Slug,
		DesiredState:     row.DesiredState,
		Config:           row.Config,
		LiveDeploymentID: row.LiveDeploymentID,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
