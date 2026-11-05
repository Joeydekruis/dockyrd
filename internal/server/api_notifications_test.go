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
