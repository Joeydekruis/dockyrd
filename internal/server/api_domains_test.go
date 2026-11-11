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

	recList := httptest.NewRecorder()
	reqList := httptest.NewRequest(http.MethodGet, "/api/v1/services/"+svc.ID+"/domains", nil)
	for _, c := range cookies {
		reqList.AddCookie(c)
	}
	mux.ServeHTTP(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("list domains status=%d %s", recList.Code, recList.Body.String())
	}
	var listed struct {
		Domains []domainResponse `json:"domains"`
	}
	_ = json.NewDecoder(recList.Body).Decode(&listed)
	var dom domainResponse
	for _, d := range listed.Domains {
		if d.IsPlatform {
			dom = d
			break
		}
	}
	if dom.ID == "" {
		t.Fatal("expected auto-provisioned platform domain")
	}
	if !strings.HasSuffix(dom.Hostname, ".203-0-113-5.sslip.io") {
		t.Fatalf("hostname %q", dom.Hostname)
	}

	domBody, _ := json.Marshal(map[string]bool{"platform": true})
	recDup := httptest.NewRecorder()
	reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/services/"+svc.ID+"/domains", bytes.NewReader(domBody))
	for _, c := range cookies {
		reqDup.AddCookie(c)
	}
	mux.ServeHTTP(recDup, reqDup)
	if recDup.Code != http.StatusConflict {
		t.Fatalf("duplicate platform domain status=%d want %d body=%s", recDup.Code, http.StatusConflict, recDup.Body.String())
	}

	recDel := httptest.NewRecorder()
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/v1/domains/"+dom.ID, nil)
	for _, c := range cookies {
		reqDel.AddCookie(c)
	}
	mux.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusNoContent {
		t.Fatalf("delete domain status=%d", recDel.Code)
	}

	recAgain := httptest.NewRecorder()
	reqAgain := httptest.NewRequest(http.MethodPost, "/api/v1/services/"+svc.ID+"/domains", bytes.NewReader(domBody))
	for _, c := range cookies {
		reqAgain.AddCookie(c)
	}
	mux.ServeHTTP(recAgain, reqAgain)
	if recAgain.Code != http.StatusCreated {
		t.Fatalf("recreate platform domain status=%d %s", recAgain.Code, recAgain.Body.String())
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
