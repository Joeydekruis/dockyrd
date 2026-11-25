package proxy

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/securitypolicy"
)

func TestCorazaDetectGolden(t *testing.T) {
	snap := securitypolicy.RouteSnapshot{
		Headers:           map[string]string{"X-Content-Type-Options": "nosniff"},
		WAFMode:           securitypolicy.WAFDetect,
		ProxyWAFInstalled: true,
		AllowWAF:          true,
		WAFPathExceptions: []string{"/api/", "/_next/"},
	}
	raw, err := Render(RenderInput{
		ACMEEmail: "ops@example.com",
		AppRoutes: []AppRoute{{
			Hostname:     "app.example.com",
			UpstreamDial: "dy-app:3000",
			Security:     &snap,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	for _, frag := range []string{
		`"handler": "coraza_waf"`,
		`"load_owasp_crs": true`,
		"SecRuleEngine DetectionOnly",
		`@beginsWith /api/`,
		`@beginsWith /_next/`,
	} {
		if !strings.Contains(s, frag) {
			t.Fatalf("missing %q\n%s", frag, s)
		}
	}
	if strings.Contains(s, "SecRuleEngine On\n") || strings.Contains(s, `"rate_limit"`) {
		t.Fatal("detect must not set On or emit rate_limit")
	}
	assertHandlerOrder(t, raw, "encode", "coraza_waf", "reverse_proxy")
}

func TestCorazaBlockGolden(t *testing.T) {
	snap := securitypolicy.RouteSnapshot{
		Headers:           map[string]string{"X-Content-Type-Options": "nosniff"},
		WAFMode:           securitypolicy.WAFBlock,
		ProxyWAFInstalled: true,
		AllowWAF:          true,
		BasicUser:         "ops",
		BasicHash:         "$2a$10$abcdefghijklmnopqrstuv",
	}
	raw, err := Render(RenderInput{
		ACMEEmail: "ops@example.com",
		AppRoutes: []AppRoute{{
			Hostname:     "block.example.com",
			UpstreamDial: "dy-block:8080",
			Security:     &snap,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, "SecRuleEngine On") {
		t.Fatal("block mode needs SecRuleEngine On")
	}
	if strings.Contains(s, "DetectionOnly") {
		t.Fatal("block must not use DetectionOnly")
	}
	assertHandlerOrder(t, raw, "encode", "coraza_waf", "authentication")
}

func TestCorazaOffOrMissingAddonOmitsHandler(t *testing.T) {
	cases := []securitypolicy.RouteSnapshot{
		{WAFMode: securitypolicy.WAFOff, ProxyWAFInstalled: true, AllowWAF: true},
		{WAFMode: securitypolicy.WAFDetect, ProxyWAFInstalled: false, AllowWAF: true},
		{WAFMode: securitypolicy.WAFDetect, ProxyWAFInstalled: true, AllowWAF: false},
	}
	for i, snap := range cases {
		snap.Headers = map[string]string{"X-Content-Type-Options": "nosniff"}
		raw, err := Render(RenderInput{
			ACMEEmail: "ops@example.com",
			AppRoutes: []AppRoute{{
				Hostname:     "x.example.com",
				UpstreamDial: "dy-x:1",
				Security:     &snap,
			}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "coraza_waf") {
			t.Fatalf("case %d unexpectedly emitted coraza_waf", i)
		}
	}
}

func TestStaticRouteNeverGetsWAF(t *testing.T) {
	snap := securitypolicy.RouteSnapshot{
		Headers:           map[string]string{"X-Content-Type-Options": "nosniff"},
		WAFMode:           securitypolicy.WAFDetect,
		ProxyWAFInstalled: true,
		AllowWAF:          false,
	}
	raw, err := Render(RenderInput{
		ACMEEmail: "ops@example.com",
		StaticRoutes: []StaticRoute{{
			Hostname: "static.example.com",
			Root:     "/var/lib/dockyrd/sites/s/current",
			Security: &snap,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "coraza_waf") {
		t.Fatal("static must never emit coraza_waf")
	}
}

func TestWAFFixtureDetectOnlyNotOn(t *testing.T) {
	// §10: Next.js API / JSON body FP shapes — detect render must stay DetectionOnly.
	dirs := buildCorazaDirectives("DetectionOnly", []string{"/api/"})
	if !strings.Contains(dirs, "SecRuleEngine DetectionOnly") {
		t.Fatal(dirs)
	}
	if strings.Contains(dirs, "SecRuleEngine On") {
		t.Fatal("fixture detect-only must not set On")
	}
}

func TestSanitizeDirectivePathRejectsEscapePayloads(t *testing.T) {
	for _, p := range []string{`/x\" …`, `/x\`, "/x\"", "/api/../x"} {
		if got := sanitizeDirectivePath(p); got != "" {
			t.Fatalf("sanitize(%q)=%q want empty", p, got)
		}
	}
	dirs := buildCorazaDirectives("DetectionOnly", []string{`/evil\"`, "/ok/"})
	if strings.Contains(dirs, `/evil`) {
		t.Fatalf("escape payload leaked into directives:\n%s", dirs)
	}
	if !strings.Contains(dirs, `@beginsWith /ok/`) {
		t.Fatalf("valid exception missing:\n%s", dirs)
	}
}

func assertHandlerOrder(t *testing.T, raw []byte, names ...string) {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	apps := doc["apps"].(map[string]any)
	httpApp := apps["http"].(map[string]any)
	servers := httpApp["servers"].(map[string]any)
	https := servers["dy_https"].(map[string]any)
	routes := https["routes"].([]any)
	var handles []any
	for _, r := range routes {
		rm := r.(map[string]any)
		handles = rm["handle"].([]any)
		break
	}
	idx := make([]int, len(names))
	for i := range idx {
		idx[i] = -1
	}
	for hi, h := range handles {
		hm, ok := h.(map[string]any)
		if !ok {
			continue
		}
		name, _ := hm["handler"].(string)
		for i, want := range names {
			if name == want && idx[i] < 0 {
				idx[i] = hi
			}
		}
	}
	for i, want := range names {
		if idx[i] < 0 {
			t.Fatalf("handler %q not found", want)
		}
		if i > 0 && idx[i] <= idx[i-1] {
			t.Fatalf("order: %q at %d should follow %q at %d", want, idx[i], names[i-1], idx[i-1])
		}
	}
}
