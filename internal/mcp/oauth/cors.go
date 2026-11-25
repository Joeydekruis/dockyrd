package oauth

import "net/http"

func setPublicCORS(w http.ResponseWriter, methods string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", methods)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// withDiscoveryCORS wraps public OAuth metadata handlers (RFC 9728 / Claude browser discovery).
func withDiscoveryCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setPublicCORS(w, "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// withRegisterCORS wraps dynamic client registration for Claude connector UI (cross-origin POST).
func withRegisterCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setPublicCORS(w, "POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}
