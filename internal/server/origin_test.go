package server

import (
	"net/http/httptest"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
)

func TestAllowedWebSocketOrigin(t *testing.T) {
	cfg := config.Config{}
	r := httptest.NewRequest("GET", "https://panel.example/ws", nil)
	r.Host = "panel.example"
	r.Header.Set("Origin", "https://panel.example")
	if !allowedWebSocketOrigin(r, cfg) {
		t.Fatal("expected match")
	}
	r.Header.Set("Origin", "https://evil.example")
	if allowedWebSocketOrigin(r, cfg) {
		t.Fatal("expected reject")
	}
}
