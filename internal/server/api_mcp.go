package server

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	mcpsdkauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	mcpserver "github.com/Joeydekruis/dockyrd/internal/mcp"
	mcpsoaut "github.com/Joeydekruis/dockyrd/internal/mcp/oauth"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/instance"
)

func (d apiDeps) registerMCPRoutes(mux *http.ServeMux) {
	if d.q == nil {
		return
	}
	panelDisabled := func(ctx context.Context) bool {
		panel, err := instance.LoadPanel(ctx, d.q)
		if err != nil {
			return false
		}
		return panel.MCPDisabled
	}
	oauth := &mcpsoaut.Server{
		Cfg: d.cfg,
		Q:   d.q,
		SessionUser: func(r *http.Request) (sqlc.GetUserByIDRow, bool) {
			u, err := d.sessionUserFromRequest(r)
			return u, err == nil
		},
		LoadPanel: func(ctx context.Context) (instance.Panel, error) {
			return instance.LoadPanel(ctx, d.q)
		},
		AllowAPIToken: func(ctx context.Context) bool {
			panel, err := instance.LoadPanel(ctx, d.q)
			if err != nil {
				return false
			}
			return panel.MCPAllowAPIToken
		},
	}
	oauth.Register(mux)

	sdkServer := mcp.NewServer(&mcp.Implementation{
		Name:    "dockyrd",
		Version: BuildVersion(),
	}, nil)
	backend := mcpserver.Backend{Cfg: d.cfg, Q: d.q, Docker: d.docker}
	backend.Register(sdkServer)

	stream := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return sdkServer
	}, &mcp.StreamableHTTPOptions{Stateless: true})

	verifier := func(ctx context.Context, token string, req *http.Request) (*mcpsdkauth.TokenInfo, error) {
		scopes, userID, exp, apiTok, err := oauth.VerifyBearerToken(ctx, token, req)
		if err != nil {
			return nil, mcpsdkauth.ErrInvalidToken
		}
		ti := &mcpsdkauth.TokenInfo{
			Scopes:     scopes,
			Expiration: exp,
			UserID:     userID,
		}
		if apiTok != nil {
			ti.Extra = map[string]any{"api_token": apiTok}
		}
		return ti, nil
	}

	mux.Handle("POST /mcp", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if panelDisabled(r.Context()) {
			http.Error(w, "mcp disabled", http.StatusNotFound)
			return
		}
		if !mcpOriginOK(r) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		panel, _ := instance.LoadPanel(r.Context(), d.q)
		origin := mcpsoaut.PanelOrigin(r, d.cfg, panel)
		prm := origin + "/.well-known/oauth-protected-resource/mcp"
		handler := mcpsdkauth.RequireBearerToken(verifier, &mcpsdkauth.RequireBearerTokenOptions{
			ResourceMetadataURL: prm,
		})(stream)
		handler.ServeHTTP(w, r)
	}))
}

func mcpOriginOK(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := r.Host
	if h := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); h != "" {
		host = h
	}
	return strings.EqualFold(u.Host, host)
}
