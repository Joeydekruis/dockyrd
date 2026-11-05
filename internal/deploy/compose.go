package deploy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/compose"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

func (rt *Runtime) runComposeDeployment(ctx context.Context, deploymentID string, dep sqlc.GetDeploymentByIDRow) error {
	ctxRow, err := rt.Q.GetServiceDeployContext(ctx, dep.ServiceID)
	if err != nil {
		return err
	}
	cfg, err := compose.Parse(appconfigRaw(ctxRow.ServiceConfig))
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	if rt.Docker == nil {
		return rt.fail(ctx, deploymentID, fmt.Errorf("docker unavailable"))
	}

	now := time.Now().Unix()
	if err := rt.Q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status:     "starting",
		Error:      sql.NullString{},
		StartedAt:  sql.NullInt64{Int64: now, Valid: true},
		FinishedAt: sql.NullInt64{},
		ID:         deploymentID,
	}); err != nil {
		return err
	}

	stackDir := compose.StackDataDir(rt.Cfg.DataDir, dep.ServiceID)
	releaseDir, inputPath, err := compose.PrepareReleaseDir(stackDir, deploymentID, cfg.Source.YAML)
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	logPath := filepath.Join(releaseDir, "deploy.log")
	_ = rt.Q.SetDeploymentReleasePath(ctx, sqlc.SetDeploymentReleasePathParams{
		ReleasePath: sql.NullString{String: releaseDir, Valid: true},
		ID:          deploymentID,
	})
	_ = rt.Q.SetDeploymentBuildLogPath(ctx, sqlc.SetDeploymentBuildLogPathParams{
		BuildLogPath: sql.NullString{String: logPath, Valid: true},
		ID:           deploymentID,
	})

	envLines, err := rt.runtimeEnv(ctx, dep.ServiceID)
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}
	if _, err := compose.WriteEnvFile(releaseDir, envLines); err != nil {
		return rt.fail(ctx, deploymentID, err)
	}

	allowPub := publishedAllowSet(cfg.PublishedPortsAllowed)
	mutIn := compose.MutateInput{
		ServiceID:      dep.ServiceID,
		DeploymentID:   deploymentID,
		ServerID:       ctxRow.ServerID,
		ProjectID:      ctxRow.ProjectID,
		StackDataDir:   stackDir,
		AllowPublished: allowPub,
	}

	runner := compose.CLIRunner{}
	deployCtx, cancel := compose.WithTimeout(ctx, cfg.WaitTimeoutSec)
	defer cancel()

	appendLog := func(line string) {
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return
		}
		_, _ = fmt.Fprintf(f, "%s\n", line)
		_ = f.Close()
	}

	doc, issues, err := compose.NormalizePipeline(deployCtx, runner, cfg, mutIn, releaseDir, inputPath, envLines, false)
	if err != nil {
		msg := formatComposeIssues(issues, err)
		appendLog(msg)
		return rt.fail(ctx, deploymentID, fmt.Errorf("%s", msg))
	}
	dockyardFile, err := compose.WriteDockyardJSON(releaseDir, doc)
	if err != nil {
		return rt.fail(ctx, deploymentID, err)
	}

	if err := runner.Pull(deployCtx, cfg.ProjectName, releaseDir, dockyardFile, envLines, appendLog); err != nil {
		appendLog(err.Error())
		return rt.fail(ctx, deploymentID, err)
	}
	if err := runner.Up(deployCtx, cfg.ProjectName, releaseDir, dockyardFile, envLines, cfg.WaitTimeoutSec, appendLog); err != nil {
		appendLog(err.Error())
		return rt.fail(ctx, deploymentID, err)
	}

	netName := compose.StackNetworkName(cfg.ProjectName)
	if err := rt.Docker.ConnectProxyToStackNetwork(deployCtx, netName); err != nil {
		appendLog("proxy network: " + err.Error())
		return rt.fail(ctx, deploymentID, err)
	}
	if err := proxy.SyncFromDB(deployCtx, rt.Q, rt.Cfg); err != nil {
		appendLog("proxy sync: " + err.Error())
		return rt.fail(ctx, deploymentID, err)
	}

	names := compose.ListServiceNames(doc)
	cfg.LastServices = names
	rawCfg, _ := json.Marshal(cfg)
	_ = rt.Q.UpdateServiceConfig(deployCtx, sqlc.UpdateServiceConfigParams{
		Config:    rawCfg,
		UpdatedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		ID:        dep.ServiceID,
	})

	finished := time.Now().Unix()
	if err := rt.Q.SupersedeLiveDeployments(deployCtx, sqlc.SupersedeLiveDeploymentsParams{
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ServiceID:  dep.ServiceID,
		ID:         deploymentID,
	}); err != nil {
		return err
	}
	if err := rt.Q.UpdateDeploymentStatus(deployCtx, sqlc.UpdateDeploymentStatusParams{
		Status:     "live",
		Error:      sql.NullString{},
		StartedAt:  sql.NullInt64{},
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ID:         deploymentID,
	}); err != nil {
		return err
	}
	if err := rt.Q.SetServiceLiveDeployment(deployCtx, sqlc.SetServiceLiveDeploymentParams{
		LiveDeploymentID: sql.NullString{String: deploymentID, Valid: true},
		UpdatedAt:        sql.NullInt64{Int64: finished, Valid: true},
		ID:               dep.ServiceID,
	}); err != nil {
		return err
	}
	_ = rt.pruneComposeReleases(stackDir, 5)
	return nil
}

// RuntimeEnv exposes runtime environment lines for compose validate API.
func (rt *Runtime) RuntimeEnv(ctx context.Context, serviceID string) ([]string, error) {
	return rt.runtimeEnv(ctx, serviceID)
}

func appconfigRaw(v interface{}) json.RawMessage {
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

func publishedAllowSet(names []string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n != "" {
			out[n] = struct{}{}
		}
	}
	return out
}

func formatComposeIssues(issues []compose.Issue, err error) string {
	var b strings.Builder
	b.WriteString(err.Error())
	for _, i := range issues {
		if i.Warning {
			continue
		}
		b.WriteString("\n- ")
		b.WriteString(i.Message)
	}
	return b.String()
}

func (rt *Runtime) pruneComposeReleases(stackDir string, keep int) error {
	releasesDir := filepath.Join(stackDir, "releases")
	entries, err := os.ReadDir(releasesDir)
	if err != nil {
		return nil
	}
	if len(entries) <= keep {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	if len(names) <= keep {
		return nil
	}
	for i := 0; i < len(names)-1; i++ {
		for j := i + 1; j < len(names); j++ {
			if names[i] < names[j] {
				names[i], names[j] = names[j], names[i]
			}
		}
	}
	for _, n := range names[keep:] {
		_ = os.RemoveAll(filepath.Join(releasesDir, n))
	}
	return nil
}
