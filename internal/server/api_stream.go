package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func registerStreamRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/stream", deps.handleStream)
}

func (d apiDeps) handleStream(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	if d.events == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "stream unavailable")
		return
	}
	topicFilter := strings.TrimSpace(r.URL.Query().Get("topics"))
	var allowed map[string]struct{}
	if topicFilter != "" {
		allowed = make(map[string]struct{})
		for _, t := range strings.Split(topicFilter, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				allowed[t] = struct{}{}
			}
		}
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := d.events.Subscribe()
	defer d.events.Unsubscribe(ch)

	fmt.Fprintf(w, "event: ready\ndata: {}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			if allowed != nil {
				var envelope struct {
					Topic string `json:"topic"`
				}
				if err := json.Unmarshal(msg, &envelope); err != nil {
					continue
				}
				if _, ok := allowed[envelope.Topic]; !ok {
					continue
				}
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}
