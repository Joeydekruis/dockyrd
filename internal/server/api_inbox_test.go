package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInboxSummaryEmpty(t *testing.T) {
	mux, cookie := testMuxWithOwnerSession(t)

	reqS := httptest.NewRequest(http.MethodGet, "/api/v1/inbox/summary", nil)
	reqS.AddCookie(cookie)
	wS := httptest.NewRecorder()
	mux.ServeHTTP(wS, reqS)
	if wS.Code != http.StatusOK {
		t.Fatalf("summary: %d %s", wS.Code, wS.Body.String())
	}
	var summary struct {
		Unread int64 `json:"unread"`
		Total  int64 `json:"total"`
	}
	if err := json.Unmarshal(wS.Body.Bytes(), &summary); err != nil {
		t.Fatal(err)
	}
	if summary.Unread != 0 || summary.Total != 0 {
		t.Fatalf("expected empty inbox, got unread=%d total=%d", summary.Unread, summary.Total)
	}

	reqL := httptest.NewRequest(http.MethodGet, "/api/v1/inbox/notifications", nil)
	reqL.AddCookie(cookie)
	wL := httptest.NewRecorder()
	mux.ServeHTTP(wL, reqL)
	if wL.Code != http.StatusOK {
		t.Fatalf("list: %d %s", wL.Code, wL.Body.String())
	}
}
