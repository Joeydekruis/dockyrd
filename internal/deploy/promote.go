package deploy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

// PromoteRelease rolls back/promotes a prior deployment (static symlink or app container).
func (rt *Runtime) PromoteRelease(ctx context.Context, serviceID, deploymentID string) error {
	svc, err := rt.Q.GetServiceByID(ctx, serviceID)
	if err != nil {
		return err
	}
	switch svc.Kind {
	case "static_site":
		return rt.PromoteStaticRelease(ctx, serviceID, deploymentID)
	case "app":
		return rt.promoteAppRelease(ctx, svc, deploymentID)
	default:
		return fmt.Errorf("promote not supported for kind %q", svc.Kind)
	}
}

func (rt *Runtime) promoteAppRelease(ctx context.Context, svc sqlc.Service, deploymentID string) error {
	dep, err := rt.Q.GetDeploymentByID(ctx, deploymentID)
	if err != nil {
		return err
	}
	if dep.ServiceID != svc.ID {
		return errors.New("deployment does not belong to service")
	}
	if !dep.ImageRef.Valid || dep.ImageRef.String == "" {
		return errors.New("deployment has no image ref")
	}
	ctxRow, err := rt.Q.GetServiceDeployContext(ctx, svc.ID)
	if err != nil {
		return err
	}
	appCfg, err := appconfig.Parse(appconfig.RawConfig(ctxRow.ServiceConfig))
	if err != nil {
		return err
	}
	if rt.Docker == nil {
		return errors.New("docker unavailable")
	}
	imageRef := dep.ImageRef.String
	if err := rt.Docker.PullImage(ctx, imageRef); err != nil {
		return fmt.Errorf("pull: %w", err)
	}
	env, err := rt.runtimeEnv(ctx, svc.ID)
	if err != nil {
		return err
	}
	containerName := docker.AppContainerName(ctxRow.ServiceSlug, docker.DeployShort(deploymentID))
	identity := docker.Identity{
		ServerID:     ctxRow.ServerID,
		ProjectID:    ctxRow.ProjectID,
		ServiceID:    svc.ID,
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
		return err
	}
	if err := rt.Docker.WaitTCPHealthy(ctx, containerID, appCfg.Port, healthTimeout); err != nil {
		_ = rt.Docker.StopAndRemove(ctx, containerID, appCfg.StopGrace)
		return err
	}
	finished := time.Now().Unix()
	if err := rt.Q.SupersedeLiveDeployments(ctx, sqlc.SupersedeLiveDeploymentsParams{
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ServiceID:  svc.ID,
		ID:         deploymentID,
	}); err != nil {
		return err
	}
	if err := rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status:     "live",
		Error:      sql.NullString{},
		StartedAt:  sql.NullInt64{},
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ID:         deploymentID,
	}); err != nil {
		return err
	}
	if err := rt.Q.SetServiceLiveDeployment(ctx, sqlc.SetServiceLiveDeploymentParams{
		LiveDeploymentID: sql.NullString{String: deploymentID, Valid: true},
		UpdatedAt:        sql.NullInt64{Int64: finished, Valid: true},
		ID:               svc.ID,
	}); err != nil {
		return err
	}
	rt.stopSupersededContainers(ctx, svc.ID, deploymentID, appCfg.StopGrace)
	return proxy.SyncFromDB(ctx, rt.Q, rt.Cfg)
}
