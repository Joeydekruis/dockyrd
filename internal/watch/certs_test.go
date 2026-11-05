package watch

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"
)

func TestHostnameMatchesCert(t *testing.T) {
	leaf := &x509.Certificate{
		DNSNames: []string{"app.example.com"},
		Subject:  pkix.Name{CommonName: "app.example.com"},
	}
	if !hostnameMatchesCert("app.example.com", leaf) {
		t.Fatal("expected match")
	}
	if hostnameMatchesCert("other.example.com", leaf) {
		t.Fatal("expected mismatch")
	}
}
