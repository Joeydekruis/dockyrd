package docker

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProxyBootstrapHTTPOnly(t *testing.T) {
	raw, err := proxyBootstrapJSON()
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if strings.Contains(s, `":443"`) {
		t.Fatalf("bootstrap must not listen on :443: %s", s)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
}
