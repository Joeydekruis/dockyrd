package proxy

import (
	"fmt"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/securitypolicy"
)

// Coraza audit log path inside the dy-proxy container (mounted from host CaddyDir).
const corazaAuditLogPath = "/var/lib/dockyrd/caddy/coraza-audit.log"

// corazaWAFHandler builds the Caddy JSON coraza_waf handler when enabled.
// Order in the handle array (after encode, before access) is the product contract;
// proxybuild images may also set global `order coraza_waf first`.
func corazaWAFHandler(snap securitypolicy.RouteSnapshot) map[string]any {
	if !snap.ProxyWAFInstalled || !snap.AllowWAF {
		return nil
	}
	mode := snap.WAFMode
	if mode != securitypolicy.WAFDetect && mode != securitypolicy.WAFBlock {
		return nil
	}
	engine := "DetectionOnly"
	if mode == securitypolicy.WAFBlock {
		engine = "On"
	}
	return map[string]any{
		"handler":        "coraza_waf",
		"load_owasp_crs": true,
		"directives":     buildCorazaDirectives(engine, snap.WAFPathExceptions),
	}
}

func buildCorazaDirectives(secRuleEngine string, pathExceptions []string) string {
	var b strings.Builder
	b.WriteString("Include @coraza.conf-recommended\n")
	b.WriteString("Include @crs-setup.conf.example\n")
	b.WriteString("Include @owasp_crs/*.conf\n")
	fmt.Fprintf(&b, "SecRuleEngine %s\n", secRuleEngine)
	b.WriteString("SecAuditEngine RelevantOnly\n")
	b.WriteString("SecAuditLog " + corazaAuditLogPath + "\n")
	b.WriteString("SecAuditLogFormat JSON\n")
	b.WriteString("SecAuditLogType Serial\n")
	b.WriteString("SecAuditLogParts ABFHZ\n")
	for i, prefix := range pathExceptions {
		prefix = sanitizeDirectivePath(prefix)
		if prefix == "" {
			continue
		}
		id := 910000 + i + 1
		fmt.Fprintf(&b,
			`SecRule REQUEST_URI "@beginsWith %s" "id:%d,phase:1,nolog,pass,ctl:ruleEngine=Off"`+"\n",
			prefix, id,
		)
	}
	return b.String()
}

func sanitizeDirectivePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || !strings.HasPrefix(p, "/") {
		return ""
	}
	if strings.ContainsAny(p, "\"\\\n\r\x00") || strings.Contains(p, "..") {
		return ""
	}
	return p
}
