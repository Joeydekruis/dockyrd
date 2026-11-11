package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Joeydekruis/dockyrd/internal/apitoken"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
)

// RunStdio starts an MCP server on stdin/stdout using DOCKYRD_URL + DOCKYRD_TOKEN.
func RunStdio(cfg config.Config) error {
	token := strings.TrimSpace(os.Getenv("DOCKYRD_TOKEN"))
	if token == "" {
		return fmt.Errorf("DOCKYRD_TOKEN required for stdio MCP")
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		return err
	}
	defer conn.Close()

	hash, err := apitoken.Hash(token)
	if err != nil {
		return err
	}
	row, err := q.GetAPITokenByHash(context.Background(), hash)
	if err != nil {
		return fmt.Errorf("invalid API token")
	}
	if !apitoken.IsUserScope(row.Scope) || row.RevokedAt.Valid {
		return fmt.Errorf("invalid API token")
	}
	if !row.CreatedBy.Valid {
		return fmt.Errorf("invalid API token")
	}

	user, err := q.GetUserByID(context.Background(), row.CreatedBy.String)
	if err != nil {
		return err
	}

	sdkServer := mcp.NewServer(&mcp.Implementation{Name: "dockyrd", Version: "stdio"}, nil)
	backend := Backend{Cfg: cfg, Q: q}
	backend.Register(sdkServer)

	// Attach principal via middleware for tool calls.
	scopes := []string{}
	if apitoken.ScopeAllows(row.Scope, apitoken.ScopeRead) {
		scopes = append(scopes, ScopeRead)
	}
	if apitoken.ScopeAllows(row.Scope, apitoken.ScopeDeploy) {
		scopes = append(scopes, ScopeDeploy)
	}
	if apitoken.ScopeAllows(row.Scope, apitoken.ScopeWrite) {
		scopes = append(scopes, ScopeWrite)
	}
	_ = user
	_ = scopes
	// Tools read auth.TokenInfoFromContext; stdio mode uses receiving middleware to inject user.
	sdkServer.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			p := Principal{User: user, Scopes: scopes, APIToken: &row}
			if row.ProjectID.Valid {
				p.ProjectID = row.ProjectID.String
			}
			ctx = WithPrincipal(ctx, p)
			return next(ctx, method, req)
		}
	})

	return sdkServer.Run(context.Background(), &mcp.StdioTransport{})
}
