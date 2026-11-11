package oauth

import (
	"fmt"
	"html"
	"net/http"
	"strings"
)

func renderOAuthConsent(w http.ResponseWriter, r *http.Request, clientID string, scopes []string) {
	q := r.URL.Query()
	q.Set("consent", "1")
	approveURL := r.URL.Path + "?" + q.Encode()
	client := strings.TrimSpace(clientID)
	if client == "" {
		client = "MCP client"
	}
	scopeLine := strings.Join(scopes, ", ")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><title>Dockyrd MCP</title></head>
<body style="font-family:system-ui,sans-serif;max-width:32rem;margin:3rem auto;padding:0 1rem">
<h1>Allow MCP access?</h1>
<p><strong>%s</strong> is requesting access to this Dockyrd instance.</p>
<p>Scopes: <code>%s</code></p>
<p><a href="%s" style="display:inline-block;padding:.75rem 1.25rem;background:#0b3d91;color:#fff;text-decoration:none;border-radius:.5rem">Approve</a>
&nbsp;<a href="/">Cancel</a></p>
</body></html>`, html.EscapeString(client), html.EscapeString(scopeLine), html.EscapeString(approveURL))
}
