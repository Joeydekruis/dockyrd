package deploy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/build"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/github"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
	"github.com/Joeydekruis/dockyrd/internal/staticsite"
)

// RunStaticSiteDeployment builds and publishes a static site release.
func (rt *Runtime) RunStaticSiteDeployment(ctx context.Context, deploymentID string) error {
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
	cfg, err := staticsite.Parse(appconfig.RawConfig(ctxRow.ServiceConfig))
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	cloneURL, branch, err := cfg.Source.CloneRemote()
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	authURL, _, err := github.ResolveClone(ctx, rt.Q, rt.Cfg.SecretKeyPath(), cfg.Source)
	if err == nil && authURL != "" {
		cloneURL = authURL
	}

	now := time.Now().Unix()
	if err := rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status:     "building",
		Error:      sql.NullString{},
		StartedAt:  sql.NullInt64{Int64: now, Valid: true},
		FinishedAt: sql.NullInt64{},
		ID:         deploymentID,
	}); err != nil {
		return err
	}

	workBase := build.DeploymentWorkDir(rt.Cfg.BuildsDir(), deploymentID)
	logPath := filepath.Join(workBase, "build.log")
	srcDir := filepath.Join(workBase, "src")
	defer func() { _ = os.RemoveAll(srcDir) }()

	if rt.Docker == nil {
		return rt.fail(ctx, deploymentID, errors.New("docker unavailable"))
	}

	if err := build.CheckDiskGuard(rt.Cfg.DataDir, rt.Cfg.MinFreeDiskPercent()); err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	buildCtx, cancel := context.WithTimeout(ctx, rt.Cfg.BuildTimeout())
	defer cancel()

	outputDir, buildErr := staticsite.Build(buildCtx, "docker", staticsite.BuildInput{
		CloneURL:     cloneURL,
		Branch:       branch,
		RootDir:      cfg.Source.RootDir,
		InstallCmd:   cfg.InstallCmd,
		BuildScript:  cfg.BuildScript,
		BuildCmd:     cfg.BuildCmd,
		NodeImage:    cfg.NodeImage,
		WorkDir:      rt.Cfg.BuildsDir(),
		DeploymentID: deploymentID,
		LogPath:      logPath,
		OutputDir:    cfg.OutputDir,
	})
	_ = rt.Q.SetDeploymentBuildLogPath(ctx, sqlc.SetDeploymentBuildLogPathParams{
		BuildLogPath: sql.NullString{String: logPath, Valid: true},
		ID:           deploymentID,
	})
	if buildErr != nil {
		return rt.fail(ctx, deploymentID, buildErr)
	}

	releasePath, err := staticsite.PublishRelease(outputDir, rt.Cfg.SitesDir(), ctxRow.ServiceSlug, deploymentID)
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	staticsite.PruneReleases(rt.Cfg.SitesDir(), ctxRow.ServiceSlug, 5)
	_ = rt.Q.SetDeploymentReleasePath(ctx, sqlc.SetDeploymentReleasePathParams{
		ReleasePath: sql.NullString{String: releasePath, Valid: true},
		ID:          deploymentID,
	})

	finished := time.Now().Unix()
	if err := rt.Q.SupersedeLiveDeployments(ctx, sqlc.SupersedeLiveDeploymentsParams{
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ServiceID:  dep.ServiceID,
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
		ID:               dep.ServiceID,
	}); err != nil {
		return err
	}

	if err := proxy.SyncFromDB(ctx, rt.Q, rt.Cfg); err != nil && rt.Log != nil {
		rt.Log.Warn("proxy sync after static deploy", "err", err)
	}
	if rt.Inbox != nil {
		rt.Inbox.NotifyDeployLive(ctx, dep.ServiceID, deploymentID)
	}
	return nil
}

// PromoteStaticRelease rolls back/promotes a prior static release to live.
func (rt *Runtime) PromoteStaticRelease(ctx context.Context, serviceID, deploymentID string) error {
	return rt.promoteStaticOnly(ctx, serviceID, deploymentID)
}

func (rt *Runtime) promoteStaticOnly(ctx context.Context, serviceID, deploymentID string) error {
	svc, err := rt.Q.GetServiceByID(ctx, serviceID)
	if err != nil {
		return err
	}
	if svc.Kind != "static_site" {
		return fmt.Errorf("promote only supported for static_site")
	}
	dep, err := rt.Q.GetDeploymentByID(ctx, deploymentID)
	if err != nil {
		return err
	}
	if dep.ServiceID != serviceID {
		return errors.New("deployment does not belong to service")
	}
	if !dep.ReleasePath.Valid || dep.ReleasePath.String == "" {
		return errors.New("deployment has no release")
	}

	releasePath, err := staticsite.ActivateRelease(rt.Cfg.SitesDir(), svc.Slug, deploymentID)
	if err != nil {
		return err
	}
	_ = releasePath

	finished := time.Now().Unix()
	if err := rt.Q.SupersedeLiveDeployments(ctx, sqlc.SupersedeLiveDeploymentsParams{
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ServiceID:  serviceID,
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
		ID:               serviceID,
	}); err != nil {
		return err
	}
	return proxy.SyncFromDB(ctx, rt.Q, rt.Cfg)
}
