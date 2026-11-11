package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Joeydekruis/dockyrd/internal/accesslog"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
)

// Backend holds control-plane dependencies for MCP tools.
type Backend struct {
	Cfg    config.Config
	Q      *sqlc.Queries
	Docker *docker.Client
}

func (b *Backend) principal(ctx context.Context) (Principal, error) {
	if p, err := PrincipalFromContext(ctx); err == nil {
		return p, nil
	}
	ti := auth.TokenInfoFromContext(ctx)
	if ti == nil || ti.UserID == "" {
		return Principal{}, errors.New("unauthenticated")
	}
	if b.Q == nil {
		return Principal{}, errors.New("database not ready")
	}
	user, err := b.Q.GetUserByID(ctx, ti.UserID)
	if err != nil {
		return Principal{}, err
	}
	p := Principal{
		User:   user,
		Scopes: ti.Scopes,
	}
	if ti.Extra != nil {
		if raw, ok := ti.Extra["api_token"].(*sqlc.ApiToken); ok && raw != nil {
			p.APIToken = raw
			if raw.ProjectID.Valid {
				p.ProjectID = raw.ProjectID.String
			}
		}
	}
	return p, nil
}

func (b *Backend) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "dockyrd_list_projects", Description: "List all projects on this Dockyrd server"}, b.toolListProjects)
	mcp.AddTool(server, &mcp.Tool{Name: "dockyrd_list_services", Description: "List services in a project (project_id required)"}, b.toolListServices)
	mcp.AddTool(server, &mcp.Tool{Name: "dockyrd_get_service", Description: "Get one service by id"}, b.toolGetService)
	mcp.AddTool(server, &mcp.Tool{Name: "dockyrd_list_deployments", Description: "List deployments for a service"}, b.toolListDeployments)
	mcp.AddTool(server, &mcp.Tool{Name: "dockyrd_get_deployment", Description: "Get deployment status by id"}, b.toolGetDeployment)
	mcp.AddTool(server, &mcp.Tool{Name: "dockyrd_get_service_logs", Description: "Tail runtime or access logs for a service"}, b.toolGetServiceLogs)
	mcp.AddTool(server, &mcp.Tool{Name: "dockyrd_trigger_deploy", Description: "Enqueue a manual deployment for a service"}, b.toolTriggerDeploy)
}

type listServicesParams struct {
	ProjectID string `json:"project_id" jsonschema:"Project id"`
}

func (b *Backend) toolListProjects(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
	p, err := b.principal(ctx)
	if err != nil {
		return nil, nil, err
	}
	if !p.HasScope(ScopeRead) {
		return nil, nil, errors.New("insufficient scope")
	}
	if p.APIToken != nil && p.APIToken.ProjectID.Valid {
		row, err := b.Q.GetProjectByID(ctx, p.APIToken.ProjectID.String)
		if err != nil {
			return nil, nil, err
		}
		return jsonToolResult([]any{row})
	}
	rows, err := b.Q.ListProjects(ctx)
	if err != nil {
		return nil, nil, err
	}
	return jsonToolResult(rows)
}

func (b *Backend) toolListServices(ctx context.Context, _ *mcp.CallToolRequest, args listServicesParams) (*mcp.CallToolResult, any, error) {
	p, err := b.principal(ctx)
	if err != nil {
		return nil, nil, err
	}
	if !p.HasScope(ScopeRead) {
		return nil, nil, errors.New("insufficient scope")
	}
	projectID := strings.TrimSpace(args.ProjectID)
	if projectID == "" {
		return nil, nil, errors.New("project_id required")
	}
	if err := b.assertProjectAccess(ctx, p, projectID); err != nil {
		return nil, nil, err
	}
	rows, err := b.Q.ListServicesByProject(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, redactedServiceListRow(row))
	}
	return jsonToolResult(out)
}

type idParam struct {
	ServiceID    string `json:"service_id,omitempty"`
	DeploymentID string `json:"deployment_id,omitempty"`
}

func (b *Backend) toolGetService(ctx context.Context, _ *mcp.CallToolRequest, args idParam) (*mcp.CallToolResult, any, error) {
	p, err := b.principal(ctx)
	if err != nil {
		return nil, nil, err
	}
	if !p.HasScope(ScopeRead) {
		return nil, nil, errors.New("insufficient scope")
	}
	id := strings.TrimSpace(args.ServiceID)
	if id == "" {
		return nil, nil, errors.New("service_id required")
	}
	projectID, err := b.Q.GetServiceProjectID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if err := b.assertProjectAccess(ctx, p, projectID); err != nil {
		return nil, nil, err
	}
	row, err := b.Q.GetServiceByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return jsonToolResult(redactedService(row))
}

func redactedServiceListRow(row sqlc.ListServicesByProjectRow) map[string]any {
	cfg := serviceConfigBytes(row.Config)
	display := dbkind.RedactConfig(row.Kind, cfg)
	out := map[string]any{
		"id":             row.ID,
		"environment_id": row.EnvironmentID,
		"kind":           row.Kind,
		"name":           row.Name,
		"slug":           row.Slug,
		"desired_state":  row.DesiredState,
		"config":         json.RawMessage(display),
	}
	if row.LiveDeploymentID.Valid {
		out["live_deployment_id"] = row.LiveDeploymentID.String
	}
	if row.CreatedAt.Valid {
		out["created_at"] = row.CreatedAt.Int64
	}
	if row.UpdatedAt.Valid {
		out["updated_at"] = row.UpdatedAt.Int64
	}
	return out
}

func redactedService(row sqlc.Service) map[string]any {
	cfg := serviceConfigBytes(row.Config)
	display := dbkind.RedactConfig(row.Kind, cfg)
	out := map[string]any{
		"id":             row.ID,
		"environment_id": row.EnvironmentID,
		"kind":           row.Kind,
		"name":           row.Name,
		"slug":           row.Slug,
		"desired_state":  row.DesiredState,
		"config":         json.RawMessage(display),
	}
	if row.CreatedAt.Valid {
		out["created_at"] = row.CreatedAt.Int64
	}
	if row.UpdatedAt.Valid {
		out["updated_at"] = row.UpdatedAt.Int64
	}
	return out
}

func serviceConfigBytes(v any) json.RawMessage {
	switch x := v.(type) {
	case []byte:
		return json.RawMessage(x)
	case string:
		return json.RawMessage(x)
	default:
		b, _ := json.Marshal(v)
		return b
	}
}

func (b *Backend) toolListDeployments(ctx context.Context, _ *mcp.CallToolRequest, args idParam) (*mcp.CallToolResult, any, error) {
	p, err := b.principal(ctx)
	if err != nil {
		return nil, nil, err
	}
	if !p.HasScope(ScopeRead) {
		return nil, nil, errors.New("insufficient scope")
	}
	id := strings.TrimSpace(args.ServiceID)
	if id == "" {
		return nil, nil, errors.New("service_id required")
	}
	projectID, err := b.Q.GetServiceProjectID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if err := b.assertProjectAccess(ctx, p, projectID); err != nil {
		return nil, nil, err
	}
	rows, err := b.Q.ListDeploymentsForService(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return jsonToolResult(rows)
}

func (b *Backend) toolGetDeployment(ctx context.Context, _ *mcp.CallToolRequest, args idParam) (*mcp.CallToolResult, any, error) {
	p, err := b.principal(ctx)
	if err != nil {
		return nil, nil, err
	}
	if !p.HasScope(ScopeRead) {
		return nil, nil, errors.New("insufficient scope")
	}
	id := strings.TrimSpace(args.DeploymentID)
	if id == "" {
		return nil, nil, errors.New("deployment_id required")
	}
	row, err := b.Q.GetDeploymentByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	projectID, err := b.Q.GetServiceProjectID(ctx, row.ServiceID)
	if err != nil {
		return nil, nil, err
	}
	if err := b.assertProjectAccess(ctx, p, projectID); err != nil {
		return nil, nil, err
	}
	return jsonToolResult(row)
}

type logsParams struct {
	ServiceID string `json:"service_id"`
	Stream    string `json:"stream,omitempty" jsonschema:"runtime or access"`
	Tail      int    `json:"tail,omitempty"`
}

func (b *Backend) toolGetServiceLogs(ctx context.Context, _ *mcp.CallToolRequest, args logsParams) (*mcp.CallToolResult, any, error) {
	p, err := b.principal(ctx)
	if err != nil {
		return nil, nil, err
	}
	if !p.HasScope(ScopeRead) {
		return nil, nil, errors.New("insufficient scope")
	}
	id := strings.TrimSpace(args.ServiceID)
	if id == "" {
		return nil, nil, errors.New("service_id required")
	}
	projectID, err := b.Q.GetServiceProjectID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if err := b.assertProjectAccess(ctx, p, projectID); err != nil {
		return nil, nil, err
	}
	svc, err := b.Q.GetServiceByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	stream := strings.TrimSpace(args.Stream)
	if stream == "" {
		stream = "runtime"
	}
	tail := args.Tail
	if tail <= 0 {
		tail = 200
	}
	if tail > 2000 {
		tail = 2000
	}
	var text string
	switch stream {
	case "access":
		text, err = accesslog.TailText(b.Cfg.CaddyAccessLogPath(), nil, tail)
	case "runtime":
		if svc.Kind == "static_site" {
			text = "Static sites have no container runtime logs. Use stream=access."
			err = nil
			break
		}
		cli := b.Docker
		if cli == nil {
			var errNew error
			cli, errNew = docker.NewClient()
			if errNew != nil {
				return nil, nil, errors.New("docker unavailable")
			}
			defer cli.Close()
		}
		if !svc.LiveDeploymentID.Valid {
			return nil, nil, errors.New("no live deployment")
		}
		cid, err := cli.FindLiveContainerForService(ctx, id, svc.LiveDeploymentID.String)
		if err != nil {
			return nil, nil, err
		}
		text, err = cli.ContainerLogs(ctx, cid, strconv.Itoa(tail))
	default:
		return nil, nil, fmt.Errorf("stream must be runtime or access")
	}
	if err != nil {
		return nil, nil, err
	}
	_ = svc
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, nil, nil
}

func (b *Backend) toolTriggerDeploy(ctx context.Context, _ *mcp.CallToolRequest, args idParam) (*mcp.CallToolResult, any, error) {
	p, err := b.principal(ctx)
	if err != nil {
		return nil, nil, err
	}
	if !p.HasScope(ScopeDeploy) {
		return nil, nil, errors.New("insufficient scope")
	}
	id := strings.TrimSpace(args.ServiceID)
	if id == "" {
		return nil, nil, errors.New("service_id required")
	}
	projectID, err := b.Q.GetServiceProjectID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if err := b.assertProjectAccess(ctx, p, projectID); err != nil {
		return nil, nil, err
	}
	deployID, err := jobs.EnqueueServiceDeploy(ctx, b.Q, id, "manual", p.User.ID, jobs.GitMeta{})
	if err != nil {
		return nil, nil, err
	}
	row, err := b.Q.GetDeploymentByID(ctx, deployID)
	if err != nil {
		return nil, nil, err
	}
	return jsonToolResult(row)
}

func (b *Backend) assertProjectAccess(ctx context.Context, p Principal, projectID string) error {
	if p.APIToken != nil && p.APIToken.ProjectID.Valid {
		if p.APIToken.ProjectID.String != projectID {
			return errors.New("forbidden")
		}
	}
	if p.User.Role == "viewer" && !p.HasScope(ScopeRead) {
		return errors.New("forbidden")
	}
	if _, err := b.Q.GetProjectByID(ctx, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("project not found")
		}
		return err
	}
	return nil
}

func jsonToolResult(v any) (*mcp.CallToolResult, any, error) {
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(buf)}},
	}, nil, nil
}
