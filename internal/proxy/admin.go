package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// Load posts a full config to Caddy's admin API over a unix socket.
func Load(ctx context.Context, adminSocket string, configJSON []byte) error {
	if adminSocket == "" {
		return fmt.Errorf("admin socket path required")
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", adminSocket)
		},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/load", bytes.NewReader(configJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("caddy load: %s: %s", resp.Status, stringsTrim(body))
	}
	return nil
}

func stringsTrim(b []byte) string {
	return string(bytes.TrimSpace(b))
}
