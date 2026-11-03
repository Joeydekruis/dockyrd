package reconcile

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// SyncDatabaseServices ensures postgres/redis containers match desired_state.
func SyncDatabaseServices(ctx context.Context, cfg config.Config, q *sqlc.Queries, dc *docker.Client, log *slog.Logger) {
	if dc == nil {
		return
	}
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		if log != nil {
			log.Debug("database reconcile: master key", "err", err)
		}
		return
	}
	rows, err := q.ListReconcileDatabaseServices(ctx)
	if err != nil {
		if log != nil {
			log.Warn("database reconcile list", "err", err)
		}
		return
	}
	managed, err := dc.ListManaged(ctx)
	if err != nil {
		if log != nil {
			log.Warn("database reconcile list containers", "err", err)
		}
		return
	}
	byService := map[string]docker.ManagedContainer{}
	for _, c := range managed {
		if c.Identity.DeploymentID != "" {
			continue
		}
		if c.Identity.Kind != docker.KindDatabase && c.Identity.Kind != docker.KindRedis {
			continue
		}
		byService[c.Identity.ServiceID] = c
	}

	wantRunning := map[string]sqlc.ListReconcileDatabaseServicesRow{}
	for _, row := range rows {
		if row.DesiredState == "running" {
			wantRunning[row.ID] = row
			continue
		}
		if c, ok := byService[row.ID]; ok {
			if err := dc.StopAndRemove(ctx, c.ID, 30); err != nil && log != nil {
				log.Warn("stop database service", "service", row.ID, "err", err)
			}
		}
	}

	for id, row := range wantRunning {
		if c, ok := byService[id]; ok && c.State == "running" {
			continue
		}
		if c, ok := byService[id]; ok {
			_ = dc.StopAndRemove(ctx, c.ID, 10)
		}
		if err := startDatabase(ctx, dc, key, row); err != nil && log != nil {
			log.Warn("start database service", "service", id, "err", err)
		}
	}
}

func startDatabase(ctx context.Context, dc *docker.Client, key []byte, row sqlc.ListReconcileDatabaseServicesRow) error {
	raw, err := configRaw(row.Config)
	if err != nil {
		return err
	}
	name := docker.ServiceContainerName(row.Slug)
	identity := docker.Identity{
		ServerID:  row.ServerID,
		ProjectID: row.ProjectID,
		ServiceID: row.ID,
		Kind:      docker.KindDatabase,
	}
	switch row.Kind {
	case "postgres":
		pg, pw, err := dbkind.ParsePostgres(raw, key)
		if err != nil {
			return err
		}
		image := fmt.Sprintf("postgres:%d", pg.Version)
		env := []string{
			"POSTGRES_USER=" + pg.User,
			"POSTGRES_PASSWORD=" + pw,
			"POSTGRES_DB=" + pg.Database,
		}
		_, err = dc.RunDatabase(ctx, docker.RunDatabaseInput{
			Name:         name,
			Image:        image,
			Env:          env,
			Identity:     identity,
			ProjectSlug:  row.ProjectSlug,
			ServiceSlug:  row.Slug,
			MemLimitMB:   pg.MemLimitMB,
			PublicPort:   pg.PublicPort,
			InternalPort: 5432,
			DataMount:    "/var/lib/postgresql/data",
		})
		return err
	case "redis":
		identity.Kind = docker.KindRedis
		redisCfg, pw, err := dbkind.ParseRedis(raw, key)
		if err != nil {
			return err
		}
		image := fmt.Sprintf("valkey/valkey:%d", redisCfg.Version)
		cmd := []string{"valkey-server", "--requirepass", pw}
		if redisCfg.Persistence {
			cmd = append(cmd, "--appendonly", "yes")
		}
		mem := redisCfg.MaxMemoryMB
		if mem <= 0 {
			mem = 256
		}
		_, err = dc.RunDatabase(ctx, docker.RunDatabaseInput{
			Name:         name,
			Image:        image,
			Cmd:          cmd,
			Identity:     identity,
			ProjectSlug:  row.ProjectSlug,
			ServiceSlug:  row.Slug,
			MemLimitMB:   mem,
			PublicPort:   0,
			InternalPort: 6379,
			DataMount:    "/data",
		})
		return err
	default:
		return fmt.Errorf("unsupported kind %q", row.Kind)
	}
}

func configRaw(v interface{}) (json.RawMessage, error) {
	switch x := v.(type) {
	case json.RawMessage:
		return x, nil
	case []byte:
		return json.RawMessage(x), nil
	case string:
		return json.RawMessage(x), nil
	default:
		return json.Marshal(v)
	}
}
