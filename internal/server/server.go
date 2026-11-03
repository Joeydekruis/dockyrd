package server

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
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
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handleHealth)

	var q *sqlc.Queries
	if _, err := os.Stat(cfg.DatabasePath()); err == nil {
		conn, queries, err := db.Open(cfg.DatabasePath())
		if err != nil {
			return err
		}
		defer conn.Close()
		q = queries
	}
	deps := apiDeps{cfg: cfg, q: q}
	registerAPIRoutes(mux, deps)
	mux.HandleFunc("GET /api/v1/setup/status", func(w http.ResponseWriter, r *http.Request) {
		handleSetupStatus(w, r, q)
	})

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

func mountUI(mux *http.ServeMux, root fs.FS) {
	fileServer := http.FileServer(http.FS(root))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		rel := strings.TrimPrefix(r.URL.Path, "/")
		if rel == "" {
			rel = "index.html"
		}
		if _, err := fs.Stat(root, rel); err != nil {
			f, err := root.Open("index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer f.Close()
			r.URL.Path = "/index.html"
		}
		fileServer.ServeHTTP(w, r)
	}))
}
