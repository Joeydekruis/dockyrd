package dns

import (
	"context"
	"testing"
)

func TestEvaluateARecordEmptyExpected(t *testing.T) {
	got := EvaluateARecord(context.Background(), "example.com", "")
	if got != "pending" {
		t.Fatalf("got %q want pending", got)
	}
}

func TestEvaluateARecordLoopbackExpected(t *testing.T) {
	got := EvaluateARecord(context.Background(), "example.com", "127.0.0.1")
	if got != "pending" {
		t.Fatalf("got %q want pending", got)
	}
}
