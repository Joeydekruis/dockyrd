package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestServiceDomainFlow(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{
		DataDir:    filepath.Join(dir, "data"),
		ConfigDir:  filepath.Join(dir, "etc"),
		PublicIPv4: "203.0.113.5",
	}
	initRes, err := setup.Init(cfg)
	if err != nil {
		t.Fatal(err)
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	mux := http.NewServeMux()
	deps := apiDeps{cfg: cfg, q: q}
	registerAPIRoutes(mux, deps)

	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Joey", "email": "o@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))
	cookies := recSetup.Result().Cookies()

	projBody, _ := json.Marshal(map[string]string{"name": "Demo"})
	recProj := httptest.NewRecorder()
	reqProj := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(projBody))
	for _, c := range cookies {
		reqProj.AddCookie(c)
	}
	mux.ServeHTTP(recProj, reqProj)
	var proj projectResponse
	_ = json.NewDecoder(recProj.Body).Decode(&proj)

	svcBody, _ := json.Marshal(map[string]any{
		"project_id": proj.ID, "kind": "app", "name": "API",
	})
	recSvc := httptest.NewRecorder()
	reqSvc := httptest.NewRequest(http.MethodPost, "/api/v1/services", bytes.NewReader(svcBody))
	for _, c := range cookies {
		reqSvc.AddCookie(c)
	}
	mux.ServeHTTP(recSvc, reqSvc)
	if recSvc.Code != http.StatusCreated {
		t.Fatalf("service status=%d %s", recSvc.Code, recSvc.Body.String())
	}
	var svc serviceResponse
	_ = json.NewDecoder(recSvc.Body).Decode(&svc)

	domBody, _ := json.Marshal(map[string]bool{"platform": true})
	recDom := httptest.NewRecorder()
	reqDom := httptest.NewRequest(http.MethodPost, "/api/v1/services/"+svc.ID+"/domains", bytes.NewReader(domBody))
	for _, c := range cookies {
		reqDom.AddCookie(c)
	}
	mux.ServeHTTP(recDom, reqDom)
	if recDom.Code != http.StatusCreated {
		t.Fatalf("domain status=%d %s", recDom.Code, recDom.Body.String())
	}
	var dom domainResponse
	_ = json.NewDecoder(recDom.Body).Decode(&dom)
	if !strings.HasSuffix(dom.Hostname, ".203-0-113-5.sslip.io") {
		t.Fatalf("hostname %q", dom.Hostname)
	}

	raw, err := proxy.Render(proxy.RenderInput{
		AppRoutes: []proxy.AppRoute{{Hostname: dom.Hostname, UpstreamDial: "dy-api-pending:80"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), dom.Hostname) {
		t.Fatal("render missing hostname")
	}
}
