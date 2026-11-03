package deploy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

const healthTimeout = 60 * time.Second

// Runtime runs image deploy jobs against Docker.
type Runtime struct {
	Cfg    config.Config
	Q      *sqlc.Queries
	Docker *docker.Client
	Log    *slog.Logger
}

// RunDeployment executes blue/green deploy for one deployment row.
func (rt *Runtime) RunDeployment(ctx context.Context, deploymentID string) error {
	dep, err := rt.Q.GetDeploymentByID(ctx, deploymentID)
	if err != nil {
		return err
	}
	if dep.Status != "queued" && dep.Status != "starting" {
		return nil
	}

	ctxRow, err := rt.Q.GetServiceDeployContext(ctx, dep.ServiceID)
	if err != nil {
		return err
	}
	appCfg, err := appconfig.Parse(appconfig.RawConfig(ctxRow.ServiceConfig))
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	imageRef, err := appCfg.ImageRef()
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}

	now := time.Now().Unix()
	if err := rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status:    "starting",
		Error:     sql.NullString{},
		StartedAt: sql.NullInt64{Int64: now, Valid: true},
		FinishedAt: sql.NullInt64{},
		ID:        deploymentID,
	}); err != nil {
		return err
	}
	if err := rt.Q.SetDeploymentImageRef(ctx, sqlc.SetDeploymentImageRefParams{
		ImageRef: sql.NullString{String: imageRef, Valid: true},
		ID:       deploymentID,
	}); err != nil {
		return err
	}

	if rt.Docker == nil {
		return rt.fail(ctx, deploymentID, errors.New("docker unavailable"))
	}

	if err := rt.Docker.PullImage(ctx, imageRef); err != nil {
		return rt.fail(ctx, deploymentID, fmt.Errorf("pull: %w", err))
	}

	env, err := rt.runtimeEnv(ctx, dep.ServiceID)
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}

	containerName := docker.AppContainerName(ctxRow.ServiceSlug, docker.DeployShort(deploymentID))
	identity := docker.Identity{
		ServerID:     ctxRow.ServerID,
		ProjectID:    ctxRow.ProjectID,
		ServiceID:    ctxRow.ServiceID,
		Kind:         docker.KindApp,
		DeploymentID: deploymentID,
	}
	containerID, err := rt.Docker.RunApp(ctx, docker.RunAppInput{
		Name:        containerName,
		Image:       imageRef,
		Port:        appCfg.Port,
		Env:         env,
		Identity:    identity,
		ProjectSlug: ctxRow.ProjectSlug,
		MemLimitMB:  appCfg.MemLimitMB,
	})
	if err != nil {
		return rt.fail(ctx, deploymentID, fmt.Errorf("run: %w", err))
	}

	if err := rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status:    "healthchecking",
		Error:     sql.NullString{},
		StartedAt: sql.NullInt64{},
		FinishedAt: sql.NullInt64{},
		ID:        deploymentID,
	}); err != nil {
		return err
	}

	if err := rt.Docker.WaitTCPHealthy(ctx, containerID, appCfg.Port, healthTimeout); err != nil {
		_ = rt.Docker.StopAndRemove(ctx, containerID, appCfg.StopGrace)
		return rt.fail(ctx, deploymentID, fmt.Errorf("health: %w", err))
	}

	finished := time.Now().Unix()
	if err := rt.Q.SupersedeLiveDeployments(ctx, sqlc.SupersedeLiveDeploymentsParams{
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ServiceID:  dep.ServiceID,
		ID:         deploymentID,
	}); err != nil {
		return err
	}
	if err := rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status:    "live",
		Error:     sql.NullString{},
		StartedAt: sql.NullInt64{},
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ID:        deploymentID,
	}); err != nil {
		return err
	}
	if err := rt.Q.SetServiceLiveDeployment(ctx, sqlc.SetServiceLiveDeploymentParams{
		LiveDeploymentID: sql.NullString{String: deploymentID, Valid: true},
		UpdatedAt:        sql.NullInt64{Int64: finished, Valid: true},
		ID:               dep.ServiceID,
	}); err != nil {
		return err
	}

	rt.stopSupersededContainers(ctx, dep.ServiceID, deploymentID, appCfg.StopGrace)
	if err := proxy.SyncFromDB(ctx, rt.Q, rt.Cfg); err != nil && rt.Log != nil {
		rt.Log.Warn("proxy sync after deploy", "err", err)
	}
	return nil
}

func (rt *Runtime) runtimeEnv(ctx context.Context, serviceID string) ([]string, error) {
	key, err := crypto.LoadMasterKey(rt.Cfg.SecretKeyPath())
	if err != nil {
		return nil, err
	}
	rows, err := rt.Q.ListVariablesForService(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.AtRuntime == 0 {
			continue
		}
		val, err := crypto.Decrypt(key, row.ValueEnc)
		if err != nil {
			return nil, err
		}
		out = append(out, fmt.Sprintf("%s=%s", row.Key, string(val)))
	}
	return out, nil
}

func (rt *Runtime) stopSupersededContainers(ctx context.Context, serviceID, liveDeploymentID string, grace int) {
	list, err := rt.Docker.ListManaged(ctx)
	if err != nil {
		return
	}
	for _, item := range list {
		if item.Identity.ServiceID != serviceID {
			continue
		}
		if item.Identity.DeploymentID == liveDeploymentID {
			continue
		}
		if item.Identity.DeploymentID == "" {
			continue
		}
		_ = rt.Docker.StopAndRemove(ctx, item.ID, grace)
	}
}

func (rt *Runtime) fail(ctx context.Context, deploymentID string, err error) error {
	msg := err.Error()
	finished := time.Now().Unix()
	_ = rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status:    "failed",
		Error:     sql.NullString{String: msg, Valid: true},
		StartedAt: sql.NullInt64{},
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ID:        deploymentID,
	})
	return err
}
