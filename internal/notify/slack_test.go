package notify

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSlackResult404DisablesChannel(t *testing.T) {
	resp := &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader("not found")),
	}
	res := classifyHTTPResult(resp, "", DeliveryMeta{})
	if !res.DisableChannel || !res.Permanent {
		t.Fatalf("404 should disable channel: %+v", res)
	}
}

func TestSlackResult429Retries(t *testing.T) {
	resp := &http.Response{
		StatusCode: 429,
		Header:     http.Header{"Retry-After": []string{"60"}},
		Body:       io.NopCloser(strings.NewReader("")),
	}
	res := classifyHTTPResult(resp, "", DeliveryMeta{})
	if res.DisableChannel || res.Permanent {
		t.Fatalf("429 should retry: %+v", res)
	}
	if res.RetryAfter <= 0 {
		t.Fatalf("expected retry-after: %+v", res)
	}
}
