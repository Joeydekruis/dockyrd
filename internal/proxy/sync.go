package proxy

import (
	"context"
	"strconv"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/compose"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/instance"
	"github.com/Joeydekruis/dockyrd/internal/securitypolicy"
	"github.com/Joeydekruis/dockyrd/internal/staticsite"
)

// SyncFromDB renders and loads Caddy config for live app and static routes.
// When docker is non-nil, app upstream hostnames (dy-*) resolve to container IPs for host-networked dy-proxy.
func SyncFromDB(ctx context.Context, q *sqlc.Queries, cfg config.Config, dock *docker.Client) error {
	inst, instErr := securitypolicy.LoadInstance(ctx, q)
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
		route := AppRoute{
			Hostname:     row.Hostname,
			UpstreamDial: upstream + ":" + strconv.Itoa(appCfg.Port),
		}
		if instErr == nil {
			site, siteErr := securitypolicy.LoadSite(ctx, q, row.ServiceID)
			if siteErr == nil {
				snap := securitypolicy.SnapshotForRoute(inst, site)
				route.Security = &snap
			}
		}
		appRoutes = append(appRoutes, route)
	}

	staticRows, err := q.ListLiveStaticProxyRoutes(ctx)
	if err != nil {
		return err
	}
	staticRoutes := make([]StaticRoute, 0, len(staticRows))
	for _, row := range staticRows {
		_ = staticsite.EnsureRelativeCurrentLink(cfg.SitesDir(), row.ServiceSlug)
		root := staticsite.CaddyFileServerRoot(row.ServiceSlug)
		if row.ReleasePath.Valid {
			if mapped, ok := staticsite.HostReleaseToCaddyRoot(cfg.SitesDir(), row.ReleasePath.String); ok {
				root = mapped
			}
		}
		var siteCfg staticsite.Config
		raw := appconfig.RawConfig(row.ServiceConfig)
		if len(raw) > 0 {
			if parsed, err := staticsite.Parse(raw); err == nil {
				siteCfg = parsed
			}
		}
		sroute := StaticRoute{
			Hostname:    row.Hostname,
			Root:        root,
			IndexNames:  staticsite.IndexFileNames(siteCfg),
			ErrorPage:   strings.TrimSpace(siteCfg.ErrorFile),
			SpaFallback: siteCfg.SpaFallback,
		}
		if instErr == nil {
			site, siteErr := securitypolicy.LoadSite(ctx, q, row.ServiceID)
			if siteErr == nil {
				snap := securitypolicy.SnapshotForRoute(inst, site)
				sroute.Security = &snap
			}
		}
		staticRoutes = append(staticRoutes, sroute)
	}

	composeRows, err := q.ListLiveComposeProxyRoutes(ctx)
	if err != nil {
		return err
	}
	for _, row := range composeRows {
		port := int(row.TargetPort.Int64)
		if !row.TargetPort.Valid || port <= 0 {
			continue
		}
		target := strings.TrimSpace(row.TargetService.String)
		if target == "" {
			continue
		}
		croute := AppRoute{
			Hostname:     row.Hostname,
			UpstreamDial: compose.UpstreamDial(row.ServiceID, target, port),
		}
		if instErr == nil {
			site, siteErr := securitypolicy.LoadSite(ctx, q, row.ServiceID)
			if siteErr == nil {
				snap := securitypolicy.SnapshotForRoute(inst, site)
				croute.Security = &snap
			}
		}
		appRoutes = append(appRoutes, croute)
	}

	if dock != nil {
		for i := range appRoutes {
			appRoutes[i].UpstreamDial = dock.ResolveHostUpstreamDial(ctx, appRoutes[i].UpstreamDial)
		}
	}

	var trusted TrustedProxyConfig
	if instErr == nil {
		trusted = TrustedFromInstance(inst)
	}

	panel, err := instance.LoadPanel(ctx, q)
	if err != nil {
		return err
	}
	acmeEmail := instance.EffectiveAcmeEmail(cfg, panel)
	if acmeEmail == "" {
		acmeEmail = "ops@localhost"
	}
	panelHost := instance.PanelHostname(instance.EffectivePublicURL(cfg, panel))
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
		AccessLog:     cfg.CaddyAccessLogPath(),
		Trusted:       trusted,
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
