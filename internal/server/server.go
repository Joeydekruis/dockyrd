package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/Joeydekruis/dockyrd/web"
)

type healthResponse struct {
	Status string `json:"status"`
}

func Serve(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handleHealth)

	ui, err := web.Assets()
	if err == nil {
		mountUI(mux, ui)
	}

	return http.ListenAndServe(addr, mux)
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
