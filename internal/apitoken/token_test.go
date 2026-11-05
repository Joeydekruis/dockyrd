package apitoken

import (
	"strings"
	"testing"
)

func TestGenerateHasPrefixAndHash(t *testing.T) {
	plain, hash, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(plain, PlainPrefix) {
		t.Fatalf("plain=%q want dyk_ prefix", plain)
	}
	got, err := Hash(plain)
	if err != nil {
		t.Fatal(err)
	}
	if len(hash) != 32 {
		t.Fatalf("hash len=%d", len(hash))
	}
	got2, err := Hash("Bearer " + plain)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(got2) {
		t.Fatal("Bearer trim mismatch")
	}
}

func TestScopeAllows(t *testing.T) {
	if !ScopeAllows(ScopeWrite, ScopeRead) {
		t.Fatal("write should allow read")
	}
	if ScopeAllows(ScopeRead, ScopeDeploy) {
		t.Fatal("read must not allow deploy")
	}
	if !ValidUserScope(ScopeDeploy) {
		t.Fatal("deploy valid")
	}
	if ValidUserScope(DeployScope("svc")) {
		t.Fatal("hook scope not a user scope")
	}
}
