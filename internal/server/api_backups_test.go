package server

import (
	"strings"
	"testing"
)

func TestRestoreUnsupportedRedisMessage(t *testing.T) {
	if !strings.Contains(restoreUnsupportedRedisMessage, "Redis") {
		t.Fatalf("expected explicit Redis message, got %q", restoreUnsupportedRedisMessage)
	}
}
