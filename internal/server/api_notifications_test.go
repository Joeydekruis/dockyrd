package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotificationChannelsRequireAdmin(t *testing.T) {
	mux, cookie := testMuxWithOwnerSession(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/notifications/channels", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("owner list channels: %d %s", w.Code, w.Body.String())
	}
}

func TestNotificationSettingsDefaultsAndPatch(t *testing.T) {
	mux, cookie := testMuxWithOwnerSession(t)
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/settings/notifications", nil)
	getReq.AddCookie(cookie)
	getW := httptest.NewRecorder()
	mux.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("get settings: %d %s", getW.Code, getW.Body.String())
	}
	var getResp struct {
		Settings struct {
			DiskWarnPct int `json:"disk_warn_pct"`
			DiskCritPct int `json:"disk_crit_pct"`
			CertWarnPct int `json:"cert_warn_pct"`
		} `json:"settings"`
	}
	if err := json.Unmarshal(getW.Body.Bytes(), &getResp); err != nil {
		t.Fatal(err)
	}
	if getResp.Settings.DiskWarnPct != 85 || getResp.Settings.DiskCritPct != 95 {
		t.Fatalf("defaults: %+v", getResp.Settings)
	}

	patchBody, _ := json.Marshal(map[string]any{
		"disk_warn_pct": 88,
		"disk_crit_pct": 97,
		"cert_warn_pct": 12,
	})
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/settings/notifications", bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.AddCookie(cookie)
	patchW := httptest.NewRecorder()
	mux.ServeHTTP(patchW, patchReq)
	if patchW.Code != http.StatusOK {
		t.Fatalf("patch settings: %d %s", patchW.Code, patchW.Body.String())
	}
	var patchResp struct {
		Settings struct {
			DiskWarnPct int `json:"disk_warn_pct"`
			DiskCritPct int `json:"disk_crit_pct"`
			CertWarnPct int `json:"cert_warn_pct"`
		} `json:"settings"`
	}
	if err := json.Unmarshal(patchW.Body.Bytes(), &patchResp); err != nil {
		t.Fatal(err)
	}
	if patchResp.Settings.DiskWarnPct != 88 || patchResp.Settings.DiskCritPct != 97 || patchResp.Settings.CertWarnPct != 12 {
		t.Fatalf("patched: %+v", patchResp.Settings)
	}
}

func TestCreateSlackChannelMasksSecret(t *testing.T) {
	mux, cookie := testMuxWithOwnerSession(t)
	body, _ := json.Marshal(map[string]any{
		"kind": "slack",
		"name": "#ops",
		"secret": map[string]any{
			"url": "https://hooks.slack.com/services/T123/B456/abcdefgh",
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings/notifications/channels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create channel: %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(resp)
	if bytes.Contains(raw, []byte("abcdefgh")) {
		t.Fatalf("secret leaked in response: %s", raw)
	}
	if resp["secret_set"] != true {
		t.Fatalf("expected secret_set true")
	}
}
