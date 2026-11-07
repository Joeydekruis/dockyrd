package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/serviceops"
	"github.com/Joeydekruis/dockyrd/internal/terminal"
)

var terminalState = struct {
	tickets  *terminal.TicketStore
	sessions *terminal.SessionRegistry
}{
	tickets:  terminal.NewTicketStore(),
	sessions: terminal.NewSessionRegistry(),
}

func newTerminalUpgrader(cfg config.Config) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return allowedWebSocketOrigin(r, cfg)
		},
	}
}

type terminalTarget struct {
	ServiceID      string `json:"service_id"`
	ServiceName    string `json:"service_name"`
	Kind           string `json:"kind"`
	ServiceKind    string `json:"service_kind"`
	ComposeService string `json:"compose_service,omitempty"`
	ContainerName  string `json:"container_name"`
}

type createTerminalSessionRequest struct {
	ServiceID      string `json:"service_id"`
	ComposeService string `json:"compose_service"`
	Mode           string `json:"mode"`
	ExecUser       string `json:"exec_user"`
}

type createTerminalSessionResponse struct {
	Ticket    string `json:"ticket"`
	ExpiresAt int64  `json:"expires_at"`
}

type wsResizeMessage struct {
	Type string `json:"type"`
	Cols uint   `json:"cols"`
	Rows uint   `json:"rows"`
}

func registerTerminalRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/terminal/targets", deps.handleTerminalTargets)
	mux.HandleFunc("POST /api/v1/terminal/sessions", deps.handleCreateTerminalSession)
	mux.HandleFunc("GET /api/v1/terminal/ws", deps.handleTerminalWebSocket)
}

func (d apiDeps) handleTerminalTargets(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	if d.docker == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "docker unavailable")
		return
	}
	ctx := r.Context()
	managed, err := d.docker.ListManaged(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "docker list error")
		return
	}
	out := make([]terminalTarget, 0)
	seen := make(map[string]struct{})
	for _, c := range managed {
		if c.State != "running" {
			continue
		}
		kind := c.Identity.Kind
		if kind == docker.KindBucket || kind == docker.KindSystem {
			continue
		}
		key := c.Identity.ServiceID + "\x00" + c.Identity.ComposeService
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		svc, err := d.q.GetServiceByID(ctx, c.Identity.ServiceID)
		if err != nil {
			continue
		}
		out = append(out, terminalTarget{
			ServiceID:      c.Identity.ServiceID,
			ServiceName:    svc.Name,
			Kind:           kind,
			ServiceKind:    svc.Kind,
			ComposeService: c.Identity.ComposeService,
			ContainerName:  c.Name,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"targets": out})
}

func (d apiDeps) handleCreateTerminalSession(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	if d.docker == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "docker unavailable")
		return
	}
	var body createTerminalSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.ServiceID = strings.TrimSpace(body.ServiceID)
	body.ComposeService = strings.TrimSpace(body.ComposeService)
	body.Mode = strings.TrimSpace(body.Mode)
	if body.Mode == "" {
		body.Mode = "shell"
	}
	execUser := strings.TrimSpace(body.ExecUser)
	if execUser == "root" && user.Role != "owner" {
		writeJSONError(w, http.StatusForbidden, "only the owner may use root in the container")
		return
	}
	ctx := r.Context()
	spec, err := d.resolveTerminalSession(ctx, user.ID, body, execUser)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	ticket, exp, err := terminalState.tickets.Issue(spec)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "ticket error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "terminal.session.start",
		TargetKind: "service",
		TargetID:   body.ServiceID,
		Meta: map[string]any{
			"mode":            body.Mode,
			"compose_service": body.ComposeService,
			"exec_user":       execUser,
		},
	})
	writeJSON(w, http.StatusOK, createTerminalSessionResponse{
		Ticket:    ticket,
		ExpiresAt: exp.Unix(),
	})
}

func (d apiDeps) resolveTerminalSession(ctx context.Context, userID string, body createTerminalSessionRequest, execUser string) (terminal.SessionSpec, error) {
	if body.ServiceID == "" {
		return terminal.SessionSpec{}, fmt.Errorf("service_id required")
	}
	svc, err := d.q.GetServiceByID(ctx, body.ServiceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return terminal.SessionSpec{}, fmt.Errorf("service not found")
		}
		return terminal.SessionSpec{}, fmt.Errorf("database error")
	}
	containerID, err := d.resolveTerminalContainer(ctx, svc, body.ComposeService)
	if err != nil {
		return terminal.SessionSpec{}, err
	}
	cmd, env, err := d.terminalCmdForMode(ctx, svc, body.Mode)
	if err != nil {
		return terminal.SessionSpec{}, err
	}
	return terminal.SessionSpec{
		UserID:         userID,
		ServiceID:      body.ServiceID,
		ContainerID:    containerID,
		ComposeService: body.ComposeService,
		Mode:           body.Mode,
		ExecUser:       execUser,
		Cmd:            cmd,
		Env:            env,
	}, nil
}

func (d apiDeps) resolveTerminalContainer(ctx context.Context, svc sqlc.Service, composeService string) (string, error) {
	if svc.Kind == "postgres" || svc.Kind == "redis" {
		return serviceops.FindDatabaseContainerID(ctx, d.docker, svc.ID)
	}
	if !svc.LiveDeploymentID.Valid {
		return "", fmt.Errorf("no live deployment for this service")
	}
	if svc.Kind == "compose" {
		if composeService == "" {
			return "", fmt.Errorf("compose_service required")
		}
		return d.docker.FindComposeServiceContainer(ctx, svc.ID, svc.LiveDeploymentID.String, composeService)
	}
	return d.docker.FindLiveContainerForService(ctx, svc.ID, svc.LiveDeploymentID.String)
}

func (d apiDeps) terminalCmdForMode(ctx context.Context, svc sqlc.Service, mode string) (cmd []string, env []string, err error) {
	switch mode {
	case "shell":
		return []string{"/bin/sh"}, nil, nil
	case "psql":
		if svc.Kind != "postgres" {
			return nil, nil, fmt.Errorf("psql shortcut is for postgres services only")
		}
		key, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
		if err != nil {
			return nil, nil, fmt.Errorf("secrets unavailable")
		}
		raw := serviceConfigJSON(svc.Config)
		pg, pw, err := dbkind.ParsePostgres(raw, key)
		if err != nil {
			return nil, nil, err
		}
		env = []string{"PGPASSWORD=" + pw}
		return []string{"psql", "-U", pg.User, "-d", pg.Database}, env, nil
	case "valkey-cli":
		if svc.Kind != "redis" {
			return nil, nil, fmt.Errorf("valkey-cli shortcut is for redis services only")
		}
		key, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
		if err != nil {
			return nil, nil, fmt.Errorf("secrets unavailable")
		}
		raw := serviceConfigJSON(svc.Config)
		_, pw, err := dbkind.ParseRedis(raw, key)
		if err != nil {
			return nil, nil, err
		}
		if pw != "" {
			env = []string{"REDISCLI_AUTH=" + pw}
		}
		return []string{"valkey-cli"}, env, nil
	default:
		return nil, nil, fmt.Errorf("unknown mode %q", mode)
	}
}

func (d apiDeps) handleTerminalWebSocket(w http.ResponseWriter, r *http.Request) {
	ticket := strings.TrimSpace(r.URL.Query().Get("ticket"))
	if ticket == "" {
		writeJSONError(w, http.StatusBadRequest, "ticket required")
		return
	}
	spec, err := terminalState.tickets.Peek(ticket)
	if err != nil {
		writeJSONError(w, http.StatusForbidden, "invalid ticket")
		return
	}
	if d.docker == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "docker unavailable")
		return
	}
	sessionID := ulid.Make().String()
	if err := terminalState.sessions.TryStart(spec.UserID, sessionID); err != nil {
		writeJSONError(w, http.StatusTooManyRequests, "too many terminal sessions")
		return
	}
	defer terminalState.sessions.End(sessionID)

	upgrader := newTerminalUpgrader(d.cfg)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	spec, err = terminalState.tickets.Redeem(ticket)
	if err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "invalid ticket"))
		return
	}

	ctx := r.Context()
	exec, err := d.docker.StartInteractiveExec(ctx, spec.ContainerID, spec.Cmd, spec.ExecUser, spec.Env)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\nCould not open a shell in this container. The image may not include /bin/sh (distroless images are unsupported).\r\n"))
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "exec failed"))
		return
	}
	defer exec.Close()

	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     spec.UserID,
		Action:     "terminal.session.connect",
		TargetKind: "service",
		TargetID:   spec.ServiceID,
		Meta: map[string]any{
			"mode":            spec.Mode,
			"compose_service": spec.ComposeService,
		},
	})

	sessionCtx, cancel := context.WithTimeout(ctx, terminal.MaxSessionDuration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	// Docker → browser
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, readErr := exec.Read(buf)
			if n > 0 {
				if werr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					return
				}
			}
			if readErr != nil {
				return
			}
		}
	}()

	// Browser → Docker + resize + idle timeout
	go func() {
		defer wg.Done()
		defer func() { _ = exec.Close() }()
		for {
			select {
			case <-sessionCtx.Done():
				return
			default:
			}
			_ = conn.SetReadDeadline(time.Now().Add(terminal.IdleTimeout))
			mt, payload, readErr := conn.ReadMessage()
			if readErr != nil {
				if errors.Is(readErr, context.Canceled) {
					return
				}
				if ne, ok := readErr.(interface{ Timeout() bool }); ok && ne.Timeout() {
					_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\nSession closed after idle timeout.\r\n"))
				}
				return
			}
			switch mt {
			case websocket.BinaryMessage:
				if _, werr := exec.Write(payload); werr != nil {
					return
				}
			case websocket.TextMessage:
				var msg wsResizeMessage
				if json.Unmarshal(payload, &msg) == nil && msg.Type == "resize" && msg.Cols > 0 && msg.Rows > 0 {
					_ = d.docker.ResizeExec(sessionCtx, exec, msg.Cols, msg.Rows)
				}
			}
		}
	}()

	wg.Wait()

	_ = d.audit.Record(context.Background(), audit.Entry{
		UserID:     spec.UserID,
		Action:     "terminal.session.end",
		TargetKind: "service",
		TargetID:   spec.ServiceID,
		Meta: map[string]any{
			"mode": spec.Mode,
		},
	})
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
