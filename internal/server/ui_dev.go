package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

const envDevUI = "dockyrd_DEV_UI"

func devUITarget() string {
	return strings.TrimRight(os.Getenv(envDevUI), "/")
}

func mountDevUIProxy(mux *http.ServeMux, target string) error {
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid %s: %q", envDevUI, target)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	origDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		origDirector(r)
		r.Host = u.Host
	}
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/.well-known/") {
			http.NotFound(w, r)
			return
		}
		proxy.ServeHTTP(w, r)
	}))
	return nil
}