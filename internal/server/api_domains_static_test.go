package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestStaticSiteDomainProvision(t *testing.T) {
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
		"project_id": proj.ID,
		"kind":       "static_site",
		"name":       "Marketing",
		"config": map[string]any{
			"source":       map[string]string{"type": "github", "repo": "acme/site", "branch": "main"},
			"buildCmd":     "npm ci && npm run build",
			"outputDir":    "dist",
			"customDomain": "clientx.example",
		},
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
		t.Fatalf("list domains status=%d", recList.Code)
	}
	var listed struct {
		Domains []domainResponse `json:"domains"`
	}
	_ = json.NewDecoder(recList.Body).Decode(&listed)
	if len(listed.Domains) < 2 {
		t.Fatalf("want platform + custom domains, got %d", len(listed.Domains))
	}
	var platform, custom int
	for _, d := range listed.Domains {
		if d.IsPlatform {
			platform++
		}
		if d.Hostname == "clientx.example" {
			custom++
			if !d.IsPrimary {
				t.Fatal("custom domain should be primary")
			}
		}
	}
	if platform != 1 || custom != 1 {
		t.Fatalf("platform=%d custom=%d", platform, custom)
	}
}
