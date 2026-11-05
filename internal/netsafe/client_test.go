package netsafe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHTTPClientBlocksLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	parsed, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	parsed.Host = "127.0.0.1:" + parsed.Port()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, parsed.String(), nil)
	client := HTTPClient(ClientOptions{})
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatal("expected dial block")
	}
}
