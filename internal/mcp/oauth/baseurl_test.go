package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/instance"
)

func TestPanelOriginUpgradesHTTPPublicURLWhenRequestIsHTTPS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://panel.example/mcp", nil)
	req.Host = "panel.example"
	req.Header.Set("X-Forwarded-Proto", "https")
	panel := instance.Panel{PublicURL: "http://panel.example"}
	got := PanelOrigin(req, config.Config{}, panel)
	if got != "https://panel.example" {
		t.Fatalf("got %q want https://panel.example", got)
	}
}
