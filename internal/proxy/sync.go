package proxy

import (
	"context"
	"strconv"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// SyncFromDB renders and loads Caddy config for all live app routes.
func SyncFromDB(ctx context.Context, q *sqlc.Queries, cfg config.Config) error {
	rows, err := q.ListLiveAppProxyRoutes(ctx)
	if err != nil {
		return err
	}
	routes := make([]AppRoute, 0, len(rows))
	for _, row := range rows {
		appCfg, err := appconfig.Parse(appconfig.RawConfig(row.ServiceConfig))
		if err != nil {
			continue
		}
		upstream := docker.AppContainerName(row.ServiceSlug, docker.DeployShort(row.DeploymentID))
		routes = append(routes, AppRoute{
			Hostname:     row.Hostname,
			UpstreamDial: upstream + ":" + strconv.Itoa(appCfg.Port),
		})
	}
	body, err := Render(RenderInput{
		ACMEEmail:   "ops@dockyard.local",
		AppRoutes:   routes,
		AdminListen: "unix/" + cfg.CaddyAdminSocket(),
	})
	if err != nil {
		return err
	}
	return Load(ctx, cfg.CaddyAdminSocket(), body)
}
