package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Joeydekruis/dockyrd/web"
)

func TestMountUI_projectDetailPlaceholder(t *testing.T) {
	root, err := web.Assets()
	if err != nil {
		t.Fatalf("Assets: %v", err)
	}
	mux := http.NewServeMux()
	mountUI(mux, root)

	for _, path := range []string{"/projects/dockyrd", "/projects/dockyrd/index.txt"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s status=%d want 200", path, rec.Code)
			}
		})
	}
}

func TestDynamicExportPlaceholder(t *testing.T) {
	cases := []struct {
		segment string
		rel     string
		want    string
	}{
		{"projects", "projects/acme", "projects/_/index.html"},
		{"projects", "projects/acme/index.txt", "projects/_/index.txt"},
		{"static-sites", "static-sites/marketing", "static-sites/_/index.html"},
		{"projects", "projects", ""},
		{"projects", "projects/_", ""},
	}
	for _, tc := range cases {
		got := dynamicExportPlaceholder(tc.rel, tc.segment)
		if got != tc.want {
			t.Errorf("segment=%q rel=%q got %q want %q", tc.segment, tc.rel, got, tc.want)
		}
	}
}

func TestMountUI_unknownPathUses404HTML(t *testing.T) {
	root, err := web.Assets()
	if err != nil {
		t.Fatalf("Assets: %v", err)
	}
	mux := http.NewServeMux()
	mountUI(mux, root)

	req := httptest.NewRequest(http.MethodGet, "/totally-unknown-path", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "404") {
		t.Fatalf("expected exported 404 page body")
	}
}
