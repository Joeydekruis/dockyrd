package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

// RepairEdgeCLI recreates dy-proxy when needed and loads Caddy routes from the control DB.
func RepairEdgeCLI(cfg config.Config) error {
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer conn.Close()

	cli, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	if err := cli.Ping(context.Background()); err != nil {
		return fmt.Errorf("docker: %w", err)
	}

	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	ctx := context.Background()
	RepairStaleLoopbackPlatformDomains(ctx, cfg, q)
	if err := proxy.SyncEdgeFromDB(ctx, cfg, q, cli, log); err != nil {
		return err
	}
	fmt.Println("Edge proxy routes loaded.")
	return nil
}
