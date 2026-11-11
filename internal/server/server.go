package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/events"
	"github.com/Joeydekruis/dockyrd/internal/inbox"
	"github.com/Joeydekruis/dockyrd/internal/notify"
	"github.com/Joeydekruis/dockyrd/internal/setup"
	"github.com/Joeydekruis/dockyrd/web"
)

type healthResponse struct {
	Status string `json:"status"`
}

type setupStatusResponse struct {
	Complete bool `json:"complete"`
}

func Serve(cfg config.Config) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handleHealth)

	hub := events.NewHub()

	var q *sqlc.Queries
	var sqlConn *sql.DB
	if _, err := os.Stat(cfg.DatabasePath()); err == nil {
		conn, queries, err := db.Open(cfg.DatabasePath())
		if err != nil {
			return err
		}
		defer conn.Close()
		q = queries
		sqlConn = conn
	}
	var dockerCli *docker.Client
	if cli, err := docker.NewClient(); err == nil {
		if err := cli.Ping(context.Background()); err == nil {
			dockerCli = cli
		} else {
			_ = cli.Close()
		}
	}
	var notifier *notify.Notifier
	var dispatcher *notify.Dispatcher
	var notifyReg notify.Registry
	if q != nil && sqlConn != nil {
		wake := make(chan struct{}, 1)
		log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
		inboxDispatch := &inbox.Dispatcher{Cfg: cfg, Q: q, Hub: hub, Log: log}
		notifier = notify.NewNotifier(sqlConn, q, cfg, hub, inboxDispatch, log, wake)
		notifyReg = notify.BuildRegistry(cfg, q, BuildVersion())
		dispatcher = notify.NewDispatcher(q, cfg, log, notifyReg, notifier, BuildVersion(), wake)
	}
	deps := apiDeps{cfg: cfg, q: q, db: sqlConn, events: hub, audit: audit.Logger{Q: q}, docker: dockerCli, notifier: notifier, notifyReg: notifyReg}
	registerAPIRoutes(mux, deps)
	mux.HandleFunc("GET /api/v1/setup/status", func(w http.ResponseWriter, r *http.Request) {
		handleSetupStatus(w, r, q)
	})

	if q != nil {
		startBackground(ctx, cfg, sqlConn, q, hub, notifier, dispatcher)
	}

	ui, err := web.Assets()
	if err == nil {
		mountUI(mux, ui)
	}

	return http.ListenAndServe(cfg.Listen, mux)
}

func handleSetupStatus(w http.ResponseWriter, _ *http.Request, q *sqlc.Queries) {
	w.Header().Set("Content-Type", "application/json")
	complete := false
	if q != nil {
		ok, err := setup.IsComplete(context.Background(), q)
		if err == nil {
			complete = ok
		}
	}
	_ = json.NewEncoder(w).Encode(setupStatusResponse{Complete: complete})
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

func spaExportFallback(rel string) string {
	for _, segment := range []string{"projects", "static-sites", "applications", "databases", "object-storage"} {
		if alt := dynamicExportPlaceholder(rel, segment); alt != "" {
			return alt
		}
	}
	switch {
	case strings.HasPrefix(rel, "server/security/addons"):
		return "server/security/addons/index.html"
	case strings.HasPrefix(rel, "server/security/policy"):
		return "server/security/policy/index.html"
	case strings.HasPrefix(rel, "server/security"):
		return "server/security/index.html"
	case strings.HasPrefix(rel, "settings/server/email"):
		return "settings/server/email/index.html"
	case strings.HasPrefix(rel, "settings/server/channels"):
		return "settings/server/channels/index.html"
	case strings.HasPrefix(rel, "settings/server/backups"):
		return "settings/server/backups/index.html"
	case strings.HasPrefix(rel, "settings/server/mcp"):
		return "settings/server/mcp/index.html"
	case strings.HasPrefix(rel, "settings/server/"):
		return "settings/server/general/index.html"
	case strings.HasPrefix(rel, "settings/account/"):
		return "settings/account/profile/index.html"
	case strings.HasPrefix(rel, "settings/instance"):
		return "settings/server/general/index.html"
	case strings.HasPrefix(rel, "settings/profile"):
		return "settings/account/profile/index.html"
	case strings.HasPrefix(rel, "settings/team"):
		return "settings/server/team/index.html"
	case strings.HasPrefix(rel, "settings/integrations"):
		return "settings/server/integrations/index.html"
	case strings.HasPrefix(rel, "settings/updates"):
		return "settings/server/updates/index.html"
	case strings.HasPrefix(rel, "reset-password"):
		return "reset-password/index.html"
	default:
		return ""
	}
}

// dynamicExportPlaceholder maps static-export placeholders (briefing §4.2), e.g.
// projects/acme → projects/_/index.html and projects/acme/index.txt → projects/_/index.txt.
func dynamicExportPlaceholder(rel, segment string) string {
	prefix := segment + "/"
	if !strings.HasPrefix(rel, prefix) {
		return ""
	}
	rest := strings.TrimPrefix(rel, prefix)
	rest = strings.TrimSuffix(rest, "/")
	if rest == "" {
		return ""
	}
	parts := strings.Split(rest, "/")
	if len(parts) == 0 || parts[0] == "_" {
		return ""
	}
	filtered := parts[:0]
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	parts = filtered
	if len(parts) == 0 || parts[0] == "_" {
		return ""
	}
	staticSegments := map[string]struct{}{
		"deployments": {},
	}
	parts[0] = "_"
	for i := 1; i < len(parts); i++ {
		if _, ok := staticSegments[parts[i]]; ok {
			continue
		}
		if strings.Contains(parts[i], ".") {
			continue
		}
		parts[i] = "_"
	}
	mapped := prefix + strings.Join(parts, "/")
	last := parts[len(parts)-1]
	if last != "index.html" && last != "index.txt" && !strings.Contains(last, ".") {
		return mapped + "/index.html"
	}
	return mapped
}

func mountUI(mux *http.ServeMux, root fs.FS) {
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/.well-known/") {
			http.NotFound(w, r)
			return
		}
		rel := strings.TrimPrefix(r.URL.Path, "/")
		rel = strings.TrimSuffix(rel, "/")
		if rel == "" {
			rel = "index.html"
		}
		if _, err := fs.Stat(root, rel); err != nil {
			indexHTML := path.Join(rel, "index.html")
			if _, err2 := fs.Stat(root, indexHTML); err2 == nil {
				rel = indexHTML
			}
		}
		if _, err := fs.Stat(root, rel); err != nil {
			if alt := spaExportFallback(rel); alt != "" {
				if _, err2 := fs.Stat(root, alt); err2 == nil {
					rel = alt
				}
			}
		}
		if _, err := fs.Stat(root, rel); err != nil {
			if _, err404 := fs.Stat(root, "404.html"); err404 == nil {
				rel = "404.html"
				w.WriteHeader(http.StatusNotFound)
			} else if _, err := fs.Stat(root, "index.html"); err != nil {
				http.NotFound(w, r)
				return
			} else {
				rel = "index.html"
			}
		}
		http.ServeFileFS(w, r, root, rel)
	}))
}
