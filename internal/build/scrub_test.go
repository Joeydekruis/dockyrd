package build

import (
	"bytes"
	"testing"
)

func TestScrubWriter(t *testing.T) {
	var buf bytes.Buffer
	w := ScrubWriter{W: &buf, Secrets: []string{"supersecret"}}
	_, _ = w.Write([]byte("token=supersecret\n"))
	if buf.String() != "token=••••\n" {
		t.Fatalf("got %q", buf.String())
	}
}
