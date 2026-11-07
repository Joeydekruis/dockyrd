package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/instance"
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
	SessionKind    string `json:"session_kind"`
	ServiceID      string `json:"service_id"`
	ComposeService string `json:"compose_service"`
	Mode           string `json:"mode"`
	ExecUser       string `json:"exec_user"`
	MfaCode        string `json:"mfa_code"`
	ProjectID      string `json:"project_id"`
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
	mux.HandleFunc("GET /api/v1/terminal/meta", deps.handleTerminalMeta)
	mux.HandleFunc("GET /api/v1/terminal/preferences", deps.handleGetTerminalPreferences)
	mux.HandleFunc("PATCH /api/v1/terminal/preferences", deps.handlePatchTerminalPreferences)
	mux.HandleFunc("POST /api/v1/terminal/sessions", deps.handleCreateTerminalSession)
	mux.HandleFunc("GET /api/v1/terminal/ws", deps.handleTerminalWebSocket)
}

func (d apiDeps) verifyUserTOTPCode(ctx context.Context, user sqlc.GetUserByIDRow, code string) error {
	if !totpEnabled(user.TotpEnabledAt) {
		return nil
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return errors.New("2fa code required")
	}
	secret, err := d.decryptUserTOTPSecret(ctx, user.ID)
	if err != nil {
		return errors.New("2fa not configured")
	}
	if !auth.ValidateTOTP(secret, code) {
		return errors.New("invalid 2fa code")
	}
	return nil
}

func (d apiDeps) handleTerminalMeta(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	panel, err := instance.LoadPanel(ctx, d.q)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	hostAllowed := panel.TerminalAllowHost
	canHost := user.Role == "owner" && hostAllowed
	writeJSON(w, http.StatusOK, map[string]any{
		"session_kinds": map[string]any{
			"container": map[string]any{"available": true},
			"host": map[string]any{
				"available":               canHost,
				"allowed_by_instance":     hostAllowed,
				"requires_mfa_if_enabled": totpEnabled(user.TotpEnabledAt),
			},
			"test": map[string]any{"available": d.docker != nil},
		},
	})
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
	var body createTerminalSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.SessionKind = strings.TrimSpace(body.SessionKind)
	if body.SessionKind == "" {
		body.SessionKind = terminal.SessionKindContainer
	}
	body.ServiceID = strings.TrimSpace(body.ServiceID)
	body.ComposeService = strings.TrimSpace(body.ComposeService)
	body.Mode = strings.TrimSpace(body.Mode)
	if body.Mode == "" {
		body.Mode = "shell"
	}
	body.MfaCode = strings.TrimSpace(body.MfaCode)
	execUser := strings.TrimSpace(body.ExecUser)
	if execUser == "root" && user.Role != "owner" {
		writeJSONError(w, http.StatusForbidden, "only the owner may use root in the container")
		return
	}
	if (body.SessionKind == terminal.SessionKindContainer ||
		body.SessionKind == terminal.SessionKindTest ||
		body.SessionKind == "") && d.docker == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "docker unavailable")
		return
	}
	ctx := r.Context()
	spec, auditMeta, err := d.resolveTerminalSession(ctx, user, body, execUser)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errTerminalForbidden) {
			status = http.StatusForbidden
		}
		if errors.Is(err, errTerminalUnauthorized) {
			status = http.StatusUnauthorized
		}
		if errors.Is(err, errTerminalUnavailable) {
			status = http.StatusNotImplemented
		}
		writeJSONError(w, status, err.Error())
		return
	}
	var ticket string
	var exp time.Time
	if spec.SessionKind == terminal.SessionKindTest && spec.ContainerID != "" {
		ticket, exp, err = terminalState.tickets.IssueWithAbandonCallback(spec, func(sp terminal.SessionSpec) {
			if d.docker == nil || sp.ContainerID == "" {
				return
			}
			_ = d.docker.StopAndRemove(context.Background(), sp.ContainerID, 10)
			_ = d.audit.Record(context.Background(), audit.Entry{
				UserID:     sp.UserID,
				Action:     "terminal.sandbox.abandoned",
				TargetKind: "project",
				TargetID:   sp.ProjectID,
				Meta: map[string]any{
					"reason": "ticket_expired",
				},
			})
		})
	} else {
		ticket, exp, err = terminalState.tickets.Issue(spec)
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "ticket error")
		return
	}
	targetKind := "service"
	targetID := body.ServiceID
	switch spec.SessionKind {
	case terminal.SessionKindHost:
		targetKind = "host"
		targetID = "server"
	case terminal.SessionKindTest:
		targetKind = "project"
		targetID = body.ProjectID
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "terminal.session.start",
		TargetKind: targetKind,
		TargetID:   targetID,
		Meta:       auditMeta,
	})
	writeJSON(w, http.StatusOK, createTerminalSessionResponse{
		Ticket:    ticket,
		ExpiresAt: exp.Unix(),
	})
}

var (
	errTerminalForbidden     = errors.New("forbidden")
	errTerminalUnauthorized  = errors.New("unauthorized")
	errTerminalUnavailable   = errors.New("not available")
)

func (d apiDeps) resolveTerminalSession(ctx context.Context, user sqlc.GetUserByIDRow, body createTerminalSessionRequest, execUser string) (terminal.SessionSpec, map[string]any, error) {
	switch body.SessionKind {
	case terminal.SessionKindHost:
		return d.resolveHostTerminalSession(ctx, user, body)
	case terminal.SessionKindTest:
		return d.resolveTestTerminalSession(ctx, user, body)
	case terminal.SessionKindContainer:
	default:
		return terminal.SessionSpec{}, nil, fmt.Errorf("unknown session_kind %q", body.SessionKind)
	}
	if body.ServiceID == "" {
		return terminal.SessionSpec{}, nil, fmt.Errorf("service_id required")
	}
	svc, err := d.q.GetServiceByID(ctx, body.ServiceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return terminal.SessionSpec{}, nil, fmt.Errorf("service not found")
		}
		return terminal.SessionSpec{}, nil, fmt.Errorf("database error")
	}
	containerID, err := d.resolveTerminalContainer(ctx, svc, body.ComposeService)
	if err != nil {
		return terminal.SessionSpec{}, nil, err
	}
	cmd, env, err := d.terminalCmdForMode(ctx, svc, body.Mode)
	if err != nil {
		return terminal.SessionSpec{}, nil, err
	}
	meta := map[string]any{
		"session_kind":    terminal.SessionKindContainer,
		"mode":            body.Mode,
		"compose_service": body.ComposeService,
		"exec_user":       execUser,
	}
	return terminal.SessionSpec{
		SessionKind:    terminal.SessionKindContainer,
		UserID:         user.ID,
		ServiceID:      body.ServiceID,
		ContainerID:    containerID,
		ComposeService: body.ComposeService,
		Mode:           body.Mode,
		ExecUser:       execUser,
		Cmd:            cmd,
		Env:            env,
	}, meta, nil
}

func (d apiDeps) resolveHostTerminalSession(ctx context.Context, user sqlc.GetUserByIDRow, body createTerminalSessionRequest) (terminal.SessionSpec, map[string]any, error) {
	if user.Role != "owner" {
		return terminal.SessionSpec{}, nil, fmt.Errorf("%w: only the owner may open a host shell", errTerminalForbidden)
	}
	panel, err := instance.LoadPanel(ctx, d.q)
	if err != nil {
		return terminal.SessionSpec{}, nil, fmt.Errorf("database error")
	}
	if !panel.TerminalAllowHost {
		return terminal.SessionSpec{}, nil, fmt.Errorf("%w: host terminal is disabled in instance settings", errTerminalForbidden)
	}
	if err := d.verifyUserTOTPCode(ctx, user, body.MfaCode); err != nil {
		return terminal.SessionSpec{}, nil, fmt.Errorf("%w: %s", errTerminalUnauthorized, err.Error())
	}
	meta := map[string]any{"session_kind": terminal.SessionKindHost}
	return terminal.SessionSpec{
		SessionKind: terminal.SessionKindHost,
		UserID:      user.ID,
		Mode:        "shell",
	}, meta, nil
}

func (d apiDeps) resolveTestTerminalSession(ctx context.Context, user sqlc.GetUserByIDRow, body createTerminalSessionRequest) (terminal.SessionSpec, map[string]any, error) {
	projectID := strings.TrimSpace(body.ProjectID)
	if projectID == "" {
		return terminal.SessionSpec{}, nil, fmt.Errorf("project_id required for test shell")
	}
	row, err := d.q.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return terminal.SessionSpec{}, nil, fmt.Errorf("project not found")
		}
		return terminal.SessionSpec{}, nil, fmt.Errorf("database error")
	}
	links, err := d.q.ListLinksForProject(ctx, projectID)
	if err != nil {
		return terminal.SessionSpec{}, nil, fmt.Errorf("database error")
	}
	linkIDs := make([]string, 0, len(links))
	for _, l := range links {
		linkIDs = append(linkIDs, l.ID)
	}
	if d.docker == nil {
		return terminal.SessionSpec{}, nil, fmt.Errorf("docker unavailable")
	}
	sandboxID, err := d.docker.RunTerminalSandbox(ctx, docker.RunTerminalSandboxInput{
		ProjectID:   projectID,
		ProjectSlug: row.Slug,
		LinksMode:   row.LinksMode,
		LinkIDs:     linkIDs,
	})
	if err != nil {
		return terminal.SessionSpec{}, nil, err
	}
	meta := map[string]any{
		"session_kind":  terminal.SessionKindTest,
		"project_id":    projectID,
		"links_mode":    row.LinksMode,
		"container_id":  sandboxID,
	}
	return terminal.SessionSpec{
		SessionKind: terminal.SessionKindTest,
		UserID:      user.ID,
		ProjectID:   projectID,
		ProjectSlug: row.Slug,
		LinksMode:   row.LinksMode,
		LinkIDs:     linkIDs,
		ContainerID: sandboxID,
		Mode:        "shell",
		Cmd:         []string{"/bin/sh"},
	}, meta, nil
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
	if spec.SessionKind == "" {
		spec.SessionKind = terminal.SessionKindContainer
	}

	targetKind := "service"
	targetID := spec.ServiceID
	switch spec.SessionKind {
	case terminal.SessionKindHost:
		targetKind = "host"
		targetID = "server"
	case terminal.SessionKindTest:
		targetKind = "project"
		targetID = spec.ProjectID
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     spec.UserID,
		Action:     "terminal.session.connect",
		TargetKind: targetKind,
		TargetID:   targetID,
		Meta: map[string]any{
			"session_kind":    spec.SessionKind,
			"mode":            spec.Mode,
			"compose_service": spec.ComposeService,
		},
	})

	sessionCtx, cancel := context.WithTimeout(ctx, terminal.MaxSessionDuration)
	defer cancel()

	type streamCloser interface {
		io.Reader
		io.Writer
		Close() error
	}
	var stream streamCloser
	var hostExec *terminal.HostPTY
	var dockerExec *docker.InteractiveExec

	switch spec.SessionKind {
	case terminal.SessionKindHost:
		var err error
		hostExec, err = terminal.StartHostShell(sessionCtx)
		if err != nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\nCould not open a host shell.\r\n"))
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "exec failed"))
			return
		}
		defer hostExec.Close()
		stream = hostExec
	default:
		if d.docker == nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\nDocker is unavailable.\r\n"))
			return
		}
		containerID := spec.ContainerID
		if spec.SessionKind == terminal.SessionKindTest {
			if containerID == "" {
				_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\nTest shell is not ready (missing sandbox).\r\n"))
				_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "sandbox missing"))
				return
			}
			defer func() {
				_ = d.docker.StopAndRemove(context.Background(), containerID, 10)
			}()
		}
		var err error
		dockerExec, err = d.docker.StartInteractiveExec(ctx, containerID, spec.Cmd, spec.ExecUser, spec.Env)
		if err != nil {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("\r\nCould not open a shell in this container. The image may not include /bin/sh (distroless images are unsupported).\r\n"))
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "exec failed"))
			return
		}
		defer dockerExec.Close()
		stream = dockerExec
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// PTY → browser
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			n, readErr := stream.Read(buf)
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

	// Browser → PTY + resize + idle timeout
	go func() {
		defer wg.Done()
		defer func() { _ = stream.Close() }()
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
				if _, werr := stream.Write(payload); werr != nil {
					return
				}
			case websocket.TextMessage:
				var msg wsResizeMessage
				if json.Unmarshal(payload, &msg) == nil && msg.Type == "resize" && msg.Cols > 0 && msg.Rows > 0 {
					if hostExec != nil {
						_ = hostExec.Resize(msg.Cols, msg.Rows)
					} else if dockerExec != nil {
						_ = d.docker.ResizeExec(sessionCtx, dockerExec, msg.Cols, msg.Rows)
					}
				}
			}
		}
	}()

	wg.Wait()

	_ = d.audit.Record(context.Background(), audit.Entry{
		UserID:     spec.UserID,
		Action:     "terminal.session.end",
		TargetKind: targetKind,
		TargetID:   targetID,
		Meta: map[string]any{
			"session_kind": spec.SessionKind,
			"mode":         spec.Mode,
		},
	})
	_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
