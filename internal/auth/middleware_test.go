package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
