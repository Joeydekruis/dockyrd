package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleSetupStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	HandleSetupStatus(rec, httptest.NewRequest(http.MethodGet, "/api/v1/setup/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var body setupStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.Complete {
		t.Fatal("expected complete=false stub")
	}
}

func TestRequireAuthPassesThrough(t *testing.T) {
	called := false
	h := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !called {
		t.Fatal("handler not called")
	}
}
