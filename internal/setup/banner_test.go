package setup

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestListenPort(t *testing.T) {
	tests := []struct {
		listen string
		want   string
	}{
		{"", "8000"},
		{":8000", "8000"},
		{"127.0.0.1:9090", "9090"},
		{"0.0.0.0:8000", "8000"},
	}
	for _, tc := range tests {
		if got := listenPort(tc.listen); got != tc.want {
			t.Errorf("listenPort(%q) = %q, want %q", tc.listen, got, tc.want)
		}
	}
}

func TestSetupBannerDirectURL(t *testing.T) {
	var buf bytes.Buffer
	exp := time.Date(2026, 11, 3, 23, 0, 0, 0, time.UTC)
	writeSetupBanner(&buf, "abc-def-ghi", exp, ":8000", func() (string, bool) {
		return "203.0.113.10", true
	})
	out := buf.String()
	if !strings.Contains(out, "http://203.0.113.10:8000/setup?token=abc-def-ghi") {
		t.Fatalf("output missing direct URL:\n%s", out)
	}
	if strings.Contains(out, "127.0.0.1") {
		t.Fatalf("should not use loopback:\n%s", out)
	}
}

func TestSetupBannerManualWhenNoIP(t *testing.T) {
	var buf bytes.Buffer
	exp := time.Date(2026, 11, 3, 23, 0, 0, 0, time.UTC)
	writeSetupBanner(&buf, "tok-123", exp, ":8000", func() (string, bool) {
		return "", false
	})
	out := buf.String()
	if !strings.Contains(out, "Setup token:  tok-123") {
		t.Fatalf("expected token line:\n%s", out)
	}
	if !strings.Contains(out, "http://<this-server-ip>:8000/setup?token=tok-123") {
		t.Fatalf("expected manual template:\n%s", out)
	}
}

func TestSetupBannerHTTPSDomainFromEnv(t *testing.T) {
	t.Setenv("dockyrd_PUBLIC_URL", "https://panel.example.com")
	var buf bytes.Buffer
	exp := time.Date(2026, 11, 3, 23, 0, 0, 0, time.UTC)
	writeSetupBanner(&buf, "abc-token", exp, ":8000", func() (string, bool) {
		return "203.0.113.10", true
	})
	out := buf.String()
	if !strings.Contains(out, "https://panel.example.com/setup?token=abc-token") {
		t.Fatalf("expected HTTPS setup URL:\n%s", out)
	}
	_ = os.Unsetenv("dockyrd_PUBLIC_URL")
}

func TestSetupBannerLoopbackBind(t *testing.T) {
	var buf bytes.Buffer
	exp := time.Date(2026, 11, 3, 23, 0, 0, 0, time.UTC)
	writeSetupBanner(&buf, "tok-123", exp, "127.0.0.1:8000", func() (string, bool) {
		return "203.0.113.10", true
	})
	out := buf.String()
	if !strings.Contains(out, "http://127.0.0.1:8000/setup?token=tok-123") {
		t.Fatalf("loopback bind should use 127.0.0.1:\n%s", out)
	}
}