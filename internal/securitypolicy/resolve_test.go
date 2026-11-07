package securitypolicy

import (
	"encoding/json"
	"testing"
)

func TestEffectiveProfileInheritUsesInstanceDefault(t *testing.T) {
	pol := Instance{DefaultProfile: ProfileStrict, Lock: false}
	site := Site{Profile: ProfileInherit}
	if got := EffectiveProfile(pol, site); got != ProfileStrict {
		t.Fatalf("got %q want strict", got)
	}
}

func TestBaselineVsStrictHSTS(t *testing.T) {
	pol := Instance{
		DefaultProfile: ProfileBaseline,
		Headers:        DefaultBaselineHeaders(),
	}
	baseline := ResolvedHeaders(pol, Site{Profile: ProfileBaseline})
	if baseline["Strict-Transport-Security"] != "max-age=15552000" {
		t.Fatalf("baseline hsts: %q", baseline["Strict-Transport-Security"])
	}
	strict := ResolvedHeaders(pol, Site{Profile: ProfileStrict})
	if strict["Strict-Transport-Security"] != "max-age=31536000" {
		t.Fatalf("strict hsts: %q", strict["Strict-Transport-Security"])
	}
	if strict["Permissions-Policy"] == "" {
		t.Fatal("expected Permissions-Policy on strict")
	}
}

func TestLockRejectsWeakerProfile(t *testing.T) {
	pol := Instance{DefaultProfile: ProfileStrict, Lock: true, Headers: DefaultBaselineHeaders()}
	site := Site{Profile: ProfileBaseline}
	err := CheckFloor(pol, site, FloorInput{Profile: ProfileBaseline}, false)
	if err == nil {
		t.Fatal("expected floor error")
	}
	if err.Field != "profile" {
		t.Fatalf("field=%q", err.Field)
	}
}

func TestLockExemptSkipsFloorReject(t *testing.T) {
	pol := Instance{DefaultProfile: ProfileStrict, Lock: true, Headers: DefaultBaselineHeaders()}
	site := Site{Profile: ProfileBaseline, LockExempt: true}
	if err := CheckFloor(pol, site, FloorInput{Profile: ProfileBaseline}, true); err != nil {
		t.Fatal(err)
	}
}

func TestLockPreventsDisablingRequiredHeader(t *testing.T) {
	h := DefaultBaselineHeaders()
	h["Strict-Transport-Security"] = HeaderEntry{Enabled: false, Value: ""}
	raw, _ := json.Marshal(h)
	pol := Instance{DefaultProfile: ProfileBaseline, Lock: true, HeadersJSON: raw}
	err := ValidatePolicyHeaders(pol)
	if err == nil {
		t.Fatal("expected header floor error")
	}
	if err.Field != "headers_json" {
		t.Fatalf("field=%q", err.Field)
	}
}

func TestDisabledHeadersOmittedFromResolved(t *testing.T) {
	h := DefaultBaselineHeaders()
	h["X-Frame-Options"] = HeaderEntry{Enabled: false, Value: ""}
	raw, _ := json.Marshal(h)
	var headers map[string]HeaderEntry
	_ = json.Unmarshal(raw, &headers)
	pol := Instance{DefaultProfile: ProfileBaseline, Headers: headers}
	out := ResolvedHeaders(pol, Site{Profile: ProfileInherit})
	if _, ok := out["X-Frame-Options"]; ok {
		t.Fatal("disabled header should be omitted")
	}
}
