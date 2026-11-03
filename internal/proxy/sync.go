package proxy

import (
	"context"
	"strconv"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/staticsite"
)

// SyncFromDB renders and loads Caddy config for live app and static routes.
func SyncFromDB(ctx context.Context, q *sqlc.Queries, cfg config.Config) error {
	appRows, err := q.ListLiveAppProxyRoutes(ctx)
	if err != nil {
		return err
	}
	appRoutes := make([]AppRoute, 0, len(appRows))
	for _, row := range appRows {
		appCfg, err := appconfig.Parse(appconfig.RawConfig(row.ServiceConfig))
		if err != nil {
			continue
		}
		upstream := docker.AppContainerName(row.ServiceSlug, docker.DeployShort(row.DeploymentID))
		appRoutes = append(appRoutes, AppRoute{
			Hostname:     row.Hostname,
			UpstreamDial: upstream + ":" + strconv.Itoa(appCfg.Port),
		})
	}

	staticRows, err := q.ListLiveStaticProxyRoutes(ctx)
	if err != nil {
		return err
	}
	staticRoutes := make([]StaticRoute, 0, len(staticRows))
	for _, row := range staticRows {
		root := staticsite.CurrentRoot(cfg.SitesDir(), row.ServiceSlug)
		staticRoutes = append(staticRoutes, StaticRoute{
			Hostname: row.Hostname,
			Root:     root,
		})
	}

	panel, err := instance.LoadPanel(ctx, q)
	if err != nil {
		return err
	}
	acmeEmail := strings.TrimSpace(panel.AcmeEmail)
	if acmeEmail == "" {
		acmeEmail = "ops@localhost"
	}
	panelHost := instance.PanelHostname(panel.PublicURL)
	panelUpstream := ""
	if panelHost != "" {
		panelUpstream = "127.0.0.1:" + cfg.ListenPort()
	}

	body, err := Render(RenderInput{
		ACMEEmail:     acmeEmail,
		PanelHost:     panelHost,
		PanelUpstream: panelUpstream,
		AppRoutes:     appRoutes,
		StaticRoutes:  staticRoutes,
		AdminListen:   adminListenPath(cfg.CaddyAdminSocket()),
	})
	if err != nil {
		return err
	}
	return Load(ctx, cfg.CaddyAdminSocket(), body)
}

func adminListenPath(socketPath string) string {
	socketPath = strings.TrimSpace(socketPath)
	socketPath = strings.TrimPrefix(socketPath, "/")
	return "unix//" + socketPath
}
