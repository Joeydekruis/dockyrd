package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

func TestServiceSecurityWAFPathExceptionsAndEvents(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Config{DataDir: filepath.Join(dir, "data"), ConfigDir: filepath.Join(dir, "etc")}
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
	registerAPIRoutes(mux, apiDeps{cfg: cfg, q: q, db: conn, audit: audit.Logger{Q: q}})
	setupBody, _ := json.Marshal(map[string]string{
		"token": initRes.Token, "name": "Owner", "email": "o@example.com", "password": "longpassword12",
	})
	recSetup := httptest.NewRecorder()
	mux.ServeHTTP(recSetup, httptest.NewRequest(http.MethodPost, "/api/v1/setup", bytes.NewReader(setupBody)))
	cookies := recSetup.Result().Cookies()

	ctx := t.Context()
	now := time.Now().UTC()
	ts := sql.NullInt64{Int64: now.Unix(), Valid: true}
	srvID := "01TESTSERVSEC000000000001"
	projID := "01TESTPROJSEC0000000001"
	envID := "01TESTENVSEC000000000001"
	svcID := "01TESTSVCSEC000000000001"
	if err := q.CreateServer(ctx, sqlc.CreateServerParams{ID: srvID, CreatedAt: ts}); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateProject(ctx, sqlc.CreateProjectParams{
		ID: projID, ServerID: srvID, Name: "Sec", Slug: "sec-waf", CreatedAt: ts, UpdatedAt: ts,
	}); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateEnvironment(ctx, sqlc.CreateEnvironmentParams{ID: envID, ProjectID: projID, Name: "production"}); err != nil {
		t.Fatal(err)
	}
	if err := q.CreateService(ctx, sqlc.CreateServiceParams{
		ID: svcID, EnvironmentID: envID, Kind: "app", Name: "web", Slug: "web",
		DesiredState: "running", Config: `{"source":{"type":"image","ref":"nginx"},"port":80}`, CreatedAt: ts, UpdatedAt: ts,
	}); err != nil {
		t.Fatal(err)
	}

	putBody, _ := json.Marshal(map[string]any{
		"profile":             "strict",
		"waf_mode":            "detect",
		"waf_path_exceptions": []string{"/api/", "/_next/"},
		"ip_allowlist":        []string{},
		"csp_report_only":     false,
		"maintenance":         false,
		"rl_mode":             "inherit",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/services/"+svcID+"/security", bytes.NewReader(putBody))
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("put=%d %s", rec.Code, rec.Body.String())
	}
	var payload serviceSecurityResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Stored.WAFMode != "detect" {
		t.Fatalf("stored waf=%q", payload.Stored.WAFMode)
	}
	if len(payload.Stored.WAFPathExceptions) != 2 {
		t.Fatalf("exceptions=%v", payload.Stored.WAFPathExceptions)
	}
	if payload.Effective.WAFMode != "detect" {
		t.Fatalf("effective waf=%q", payload.Effective.WAFMode)
	}

	badBody, _ := json.Marshal(map[string]any{
		"profile":             "strict",
		"waf_path_exceptions": []string{"../etc"},
		"ip_allowlist":        []string{},
	})
	reqBad := httptest.NewRequest(http.MethodPut, "/api/v1/services/"+svcID+"/security", bytes.NewReader(badBody))
	for _, c := range cookies {
		reqBad.AddCookie(c)
	}
	recBad := httptest.NewRecorder()
	mux.ServeHTTP(recBad, reqBad)
	if recBad.Code != http.StatusUnprocessableEntity {
		t.Fatalf("bad exceptions=%d", recBad.Code)
	}

	_, err = q.InsertSecurityEvent(ctx, sqlc.InsertSecurityEventParams{
		Ts: now.Format(time.RFC3339Nano), ServiceID: svcID, Source: "waf",
		Rule:   sql.NullString{String: "942100", Valid: true},
		Path:   sql.NullString{String: "/api/foo", Valid: true},
		Action: sql.NullString{String: "detect", Valid: true},
		Ip:     sql.NullString{String: "203.0.113.10", Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	reqEv := httptest.NewRequest(http.MethodGet, "/api/v1/services/"+svcID+"/security/events", nil)
	for _, c := range cookies {
		reqEv.AddCookie(c)
	}
	recEv := httptest.NewRecorder()
	mux.ServeHTTP(recEv, reqEv)
	if recEv.Code != http.StatusOK {
		t.Fatalf("events=%d %s", recEv.Code, recEv.Body.String())
	}
	var evBody struct {
		Events []securityEventDTO `json:"events"`
	}
	if err := json.Unmarshal(recEv.Body.Bytes(), &evBody); err != nil {
		t.Fatal(err)
	}
	if len(evBody.Events) != 1 || evBody.Events[0].Source != "waf" {
		t.Fatalf("%+v", evBody.Events)
	}
}
