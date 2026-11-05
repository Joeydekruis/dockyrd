package deploy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/build"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
	"github.com/Joeydekruis/dockyrd/internal/inbox"
	"github.com/Joeydekruis/dockyrd/internal/github"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
	"github.com/Joeydekruis/dockyrd/internal/refs"
	"github.com/Joeydekruis/dockyrd/internal/resources"
)

const healthTimeout = 60 * time.Second

// Runtime runs image deploy jobs against Docker.
type Runtime struct {
	Cfg       config.Config
	Q         *sqlc.Queries
	Docker    *docker.Client
	Log       *slog.Logger
	Events    *eventlog.Logger
	Inbox     *inbox.Dispatcher
	Resources resources.Resources
}

// RunDeployment executes deploy for app or static_site services.
func (rt *Runtime) RunDeployment(ctx context.Context, deploymentID string) error {
	dep, err := rt.Q.GetDeploymentByID(ctx, deploymentID)
	if err != nil {
		return err
	}
	if dep.Status != "queued" && dep.Status != "starting" && dep.Status != "waiting" {
		return nil
	}
	svc, err := rt.Q.GetServiceByID(ctx, dep.ServiceID)
	if err != nil {
		return err
	}
	switch svc.Kind {
	case "static_site":
		return rt.RunStaticSiteDeployment(ctx, deploymentID)
	case "app":
		return rt.runAppDeployment(ctx, deploymentID, dep)
	default:
		return rt.fail(ctx, deploymentID, fmt.Errorf("deploy not supported for kind %q", svc.Kind))
	}
}

func (rt *Runtime) runAppDeployment(ctx context.Context, deploymentID string, dep sqlc.GetDeploymentByIDRow) error {
	ctxRow, err := rt.Q.GetServiceDeployContext(ctx, dep.ServiceID)
	if err != nil {
		return err
	}
	appCfg, err := appconfig.Parse(appconfig.RawConfig(ctxRow.ServiceConfig))
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	deployShort := docker.DeployShort(deploymentID)
	imageRef, needsBuild, err := appCfg.DeploymentImageRef(ctxRow.ServiceSlug, deployShort)
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	if dep.ImageRef.Valid && strings.TrimSpace(dep.ImageRef.String) != "" {
		imageRef = strings.TrimSpace(dep.ImageRef.String)
		needsBuild = false
	}
	if appCfg.UsesGitHubActionsBuild() && needsBuild {
		if dep.Status == "waiting" {
			return nil
		}
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

	if needsBuild {
		builder := strings.TrimSpace(appCfg.Builder)
		if builder == "" {
			builder = appconfig.BuilderRailpack
		}
		isNext := strings.Contains(strings.ToLower(appCfg.Source.Repo+appCfg.Source.URL), "next")
		peak := build.ExpectedPeakHeuristic(builder, isNext)
		budget, err := rt.acquireBuildBudget(ctx, deploymentID, dep.ServiceID, peak)
		if err != nil {
			return err
		}
		if err := rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
			Status:    "building",
			Error:     sql.NullString{},
			StartedAt: sql.NullInt64{},
			FinishedAt: sql.NullInt64{},
			ID:        deploymentID,
		}); err != nil {
			return err
		}
		buildArgs, secrets, err := rt.buildArgs(ctx, dep.ServiceID)
		if err != nil {
			return rt.fail(ctx, deploymentID, err)
		}
		buildArgs = build.MergeBuildArgs(buildArgs, budget.RAMBytes, budget.SwapBytes)
		cloneURL, _, cloneErr := github.ResolveClone(ctx, rt.Q, rt.Cfg.SecretKeyPath(), appCfg.Source)
		if cloneErr != nil {
			return rt.fail(ctx, deploymentID, cloneErr)
		}
		buildCtx, cancel := context.WithTimeout(ctx, rt.Cfg.BuildTimeout())
		defer cancel()
		stopWatch := rt.startBuildWatchdog(buildCtx, dep.ServiceID, deploymentID, cancel)
		defer stopWatch()
		sampler := build.StartPeakSampler(buildCtx, rt.Resources)
		logPath, buildErr := build.RunGitAppBuild(buildCtx, build.GitAppBuildInput{
			WorkDir:            rt.Cfg.BuildsDir(),
			DataDir:            rt.Cfg.DataDir,
			MinFreeDiskPercent: rt.Cfg.MinFreeDiskPercent(),
			DeploymentID:       deploymentID,
			ServiceID:          dep.ServiceID,
			App:                appCfg,
			Tag:                imageRef,
			CloneURL:           cloneURL,
			BuildArgs:          buildArgs,
			ScrubSecrets:       secrets,
		})
		_ = rt.Q.SetDeploymentBuildLogPath(ctx, sqlc.SetDeploymentBuildLogPathParams{
			BuildLogPath: sql.NullString{String: logPath, Valid: logPath != ""},
			ID:           deploymentID,
		})
		if buildErr != nil {
			return rt.fail(ctx, deploymentID, buildErr)
		}
		rt.recordBuildMetrics(ctx, deploymentID, budget, "local", sampler)
		if err := rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
			Status:    "starting",
			Error:     sql.NullString{},
			StartedAt: sql.NullInt64{},
			FinishedAt: sql.NullInt64{},
			ID:        deploymentID,
		}); err != nil {
			return err
		}
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

	if !needsBuild {
		if err := rt.Docker.PullImage(ctx, imageRef); err != nil {
			return rt.fail(ctx, deploymentID, fmt.Errorf("pull: %w", err))
		}
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
	attachEdge, err := rt.appAttachEdge(ctx, dep.ServiceID)
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	skipProjectNet, err := rt.skipProjectNetwork(ctx, ctxRow.ProjectID)
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	containerID, err := rt.Docker.RunApp(ctx, docker.RunAppInput{
		Name:               containerName,
		Image:              imageRef,
		Port:               appCfg.Port,
		Env:                env,
		Identity:           identity,
		ProjectSlug:        ctxRow.ProjectSlug,
		MemLimitMB:         appCfg.MemLimitMB,
		AttachEdge:         attachEdge,
		SkipProjectNetwork: skipProjectNet,
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
	rt.recordDeployEvent(ctx, dep.ServiceID, deploymentID, "info", "Deploy live", "")
	if rt.Inbox != nil {
		rt.Inbox.NotifyDeployLive(ctx, dep.ServiceID, deploymentID)
	}
	if needsBuild && rt.Docker != nil {
		slug := ctxRow.ServiceSlug
		go func() {
			_ = rt.Docker.PruneServiceImages(context.Background(), slug, 5)
		}()
	}
	return nil
}

func (rt *Runtime) runtimeEnv(ctx context.Context, serviceID string) ([]string, error) {
	ctxRow, err := rt.Q.GetServiceDeployContext(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	catalog, err := rt.consumerCatalog(ctx, serviceID, ctxRow.ProjectID)
	if err != nil {
		return nil, err
	}
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
		resolved, err := refs.Resolve(string(val), refs.Catalog(catalog))
		if err != nil {
			return nil, err
		}
		out = append(out, fmt.Sprintf("%s=%s", row.Key, resolved))
	}
	return out, nil
}

func (rt *Runtime) buildArgs(ctx context.Context, serviceID string) (map[string]string, []string, error) {
	ctxRow, err := rt.Q.GetServiceDeployContext(ctx, serviceID)
	if err != nil {
		return nil, nil, err
	}
	catalog, err := rt.consumerCatalog(ctx, serviceID, ctxRow.ProjectID)
	if err != nil {
		return nil, nil, err
	}
	key, err := crypto.LoadMasterKey(rt.Cfg.SecretKeyPath())
	if err != nil {
		return nil, nil, err
	}
	rows, err := rt.Q.ListVariablesForService(ctx, serviceID)
	if err != nil {
		return nil, nil, err
	}
	args := make(map[string]string)
	secrets := make([]string, 0)
	for _, row := range rows {
		if row.AtBuild == 0 {
			continue
		}
		val, err := crypto.Decrypt(key, row.ValueEnc)
		if err != nil {
			return nil, nil, err
		}
		s, err := refs.Resolve(string(val), refs.Catalog(catalog))
		if err != nil {
			return nil, nil, err
		}
		args[row.Key] = s
		if row.IsSecret != 0 {
			secrets = append(secrets, s)
		}
	}
	return args, secrets, nil
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
	kind := ""
	if strings.Contains(strings.ToLower(msg), "out of memory") || strings.Contains(msg, "137") {
		kind = "oom"
		msg = build.FailureMessage(kind, msg)
	}
	finished := time.Now().Unix()
	dep, _ := rt.Q.GetDeploymentByID(ctx, deploymentID)
	_ = rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status:    "failed",
		Error:     sql.NullString{String: msg, Valid: true},
		StartedAt: sql.NullInt64{},
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ID:        deploymentID,
	})
	if kind != "" {
		_ = rt.Q.SetDeploymentFailureKind(ctx, sqlc.SetDeploymentFailureKindParams{
			FailureKind: sql.NullString{String: kind, Valid: true},
			ID:          deploymentID,
		})
	}
	if dep.ServiceID != "" {
		rt.recordDeployEvent(ctx, dep.ServiceID, deploymentID, "error", "Deploy failed", msg)
		if rt.Inbox != nil {
			rt.Inbox.NotifyDeployFailed(ctx, dep.ServiceID, deploymentID, msg)
		}
	}
	return err
}

func (rt *Runtime) recordDeployEvent(ctx context.Context, serviceID, deploymentID, level, title, detail string) {
	if rt.Events == nil {
		return
	}
	ctxRow, err := rt.Q.GetServiceDeployContext(ctx, serviceID)
	if err != nil {
		return
	}
	_ = rt.Events.Record(ctx, eventlog.RecordInput{
		Level:        level,
		Category:     "deploys",
		Title:        title,
		Detail:       detail,
		ProjectID:    ctxRow.ProjectID,
		ServiceID:    serviceID,
		DeploymentID: deploymentID,
	})
}
