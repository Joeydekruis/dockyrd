package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMountDevUIProxy_forwardsNonAPI(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "next:"+r.URL.Path)
	}))
	t.Cleanup(backend.Close)

	mux := http.NewServeMux()
	if err := mountDevUIProxy(mux, backend.URL); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/projects/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
	if got := rec.Body.String(); got != "next:/projects/" {
		t.Fatalf("body=%q", got)
	}
}

func TestMountDevUIProxy_skipsAPIPrefix(t *testing.T) {
	backend := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(backend.Close)

	mux := http.NewServeMux()
	if err := mountDevUIProxy(mux, backend.URL); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rec.Code)
	}
}