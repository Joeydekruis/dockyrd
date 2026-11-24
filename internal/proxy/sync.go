package proxy

import (
	"context"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// SyncFromDB renders and loads Caddy config for live app and static routes.
// When docker is non-nil, app upstream hostnames (dy-*) resolve to container IPs for host-networked dy-proxy.
func SyncFromDB(ctx context.Context, q *sqlc.Queries, cfg config.Config, dock *docker.Client) error {
	body, err := RenderFromDB(ctx, q, cfg, dock)
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