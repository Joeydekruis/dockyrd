package securitypolicy

import "testing"

func TestEffectiveWAFModeInheritStrictIsDetectNeverBlock(t *testing.T) {
	inst := Instance{DefaultProfile: ProfileStrict}
	site := Site{Profile: ProfileInherit, WAFMode: WAFInherit}
	if got := EffectiveWAFMode(inst, site); got != WAFDetect {
		t.Fatalf("strict inherit → detect, got %q", got)
	}
}

func TestEffectiveWAFModeInheritBaselineIsOff(t *testing.T) {
	inst := Instance{DefaultProfile: ProfileBaseline}
	site := Site{Profile: ProfileInherit, WAFMode: WAFInherit}
	if got := EffectiveWAFMode(inst, site); got != WAFOff {
		t.Fatalf("baseline inherit → off, got %q", got)
	}
}

func TestEffectiveWAFModeExplicitBlockWins(t *testing.T) {
	inst := Instance{DefaultProfile: ProfileStrict}
	site := Site{Profile: ProfileStrict, WAFMode: WAFBlock}
	if got := EffectiveWAFMode(inst, site); got != WAFBlock {
		t.Fatalf("explicit block, got %q", got)
	}
}

func TestEffectiveWAFModeCustomInheritIsOff(t *testing.T) {
	inst := Instance{DefaultProfile: ProfileBaseline}
	site := Site{Profile: ProfileCustom, WAFMode: WAFInherit}
	if got := EffectiveWAFMode(inst, site); got != WAFOff {
		t.Fatalf("custom inherit → off floor, got %q", got)
	}
}

func TestValidatePathExceptionsRejectsDotDot(t *testing.T) {
	_, err := ValidatePathExceptions([]string{"/api/../secret"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidatePathExceptionsRejectsEscapePayloads(t *testing.T) {
	for _, p := range []string{`/x\" id:1`, `/x\n`, `/api/\..\secret`} {
		if _, err := ValidatePathExceptions([]string{p}); err == nil {
			t.Fatalf("expected reject for %q", p)
		}
	}
}

func TestValidatePathExceptionsOK(t *testing.T) {
	out, err := ValidatePathExceptions([]string{"/api/", "/_next/"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("len=%d", len(out))
	}
}

func TestValidatePathExceptionsMaxCount(t *testing.T) {
	in := make([]string, 65)
	for i := range in {
		in[i] = "/" + string(rune('a'+i%26))
	}
	_, err := ValidatePathExceptions(in)
	if err == nil {
		t.Fatal("expected max count error")
	}
}

func TestValidateWAFMode(t *testing.T) {
	if err := ValidateWAFMode("block"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateWAFMode("nope"); err == nil {
		t.Fatal("expected error")
	}
}
