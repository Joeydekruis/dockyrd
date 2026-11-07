package reconcile

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/compose"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

const composeHealCooldown = 60 * time.Second

var lastComposeHeal = map[string]time.Time{}

// SyncComposeServices ensures compose stacks stay healthy and proxy stays attached.
func SyncComposeServices(ctx context.Context, cfg config.Config, q *sqlc.Queries, cli *docker.Client, log *slog.Logger) {
	if cli == nil || q == nil {
		return
	}
	rows, err := q.ListServicesByKind(ctx, "compose")
	if err != nil {
		if log != nil {
			log.Debug("list compose services", "err", err)
		}
		return
	}
	runner := compose.CLIRunner{}
	for _, row := range rows {
		if row.DesiredState != "running" {
			continue
		}
		if !row.LiveDeploymentID.Valid {
			continue
		}
		cfgJSON, err := compose.Parse(rawConfig(row.Config))
		if err != nil {
			continue
		}
		releaseDir := filepath.Join(compose.StackDataDir(cfg.DataDir, row.ID), "releases", row.LiveDeploymentID.String)
		composeFile := filepath.Join(releaseDir, "compose.dockyard.json")
		if _, err := os.Stat(composeFile); err != nil {
			continue
		}
		netName := compose.StackNetworkName(cfgJSON.ProjectName)
		if err := cli.ConnectProxyToStackNetwork(ctx, netName); err != nil && log != nil {
			log.Debug("compose proxy network", "service", row.ID, "err", err)
		}

		expected := cfgJSON.LastServices
		if len(expected) == 0 {
			if raw, err := os.ReadFile(composeFile); err == nil {
				var doc map[string]any
				if json.Unmarshal(raw, &doc) == nil {
					expected = compose.ListServiceNames(doc)
				}
			}
		}
		statuses, err := runner.PS(ctx, cfgJSON.ProjectName, releaseDir, composeFile)
		if err != nil {
			maybeHealCompose(ctx, row.ID, cfgJSON, releaseDir, composeFile, runner, log)
			continue
		}
		running := map[string]bool{}
		for _, st := range statuses {
			if strings.EqualFold(st.State, "running") {
				running[st.Service] = true
			}
		}
		healthy := true
		for _, name := range expected {
			if !running[name] {
				healthy = false
				break
			}
		}
		if !healthy {
			maybeHealCompose(ctx, row.ID, cfgJSON, releaseDir, composeFile, runner, log)
		}
	}
	_ = proxy.SyncFromDB(ctx, q, cfg, cli)
}

func maybeHealCompose(ctx context.Context, serviceID string, cfg compose.Config, releaseDir, composeFile string, runner compose.CLIRunner, log *slog.Logger) {
	if t, ok := lastComposeHeal[serviceID]; ok && time.Since(t) < composeHealCooldown {
		return
	}
	lastComposeHeal[serviceID] = time.Now()
	healCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.WaitTimeoutSec)*time.Second)
	defer cancel()
	if err := runner.Up(healCtx, cfg.ProjectName, releaseDir, composeFile, nil, cfg.WaitTimeoutSec, nil); err != nil && log != nil {
		log.Warn("compose heal up", "service", serviceID, "err", err)
	}
}

func rawConfig(v interface{}) json.RawMessage {
	switch x := v.(type) {
	case json.RawMessage:
		return x
	case []byte:
		return json.RawMessage(x)
	case string:
		return json.RawMessage(x)
	default:
		b, _ := json.Marshal(x)
		return b
	}
}
