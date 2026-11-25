package events

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ParsedHit is one WAF security event derived from a Coraza audit log line.
type ParsedHit struct {
	TS       time.Time
	Hostname string
	Path     string
	Rule     string
	Action   string
	IP       string
	Detail   string
}

type corazaAuditLine struct {
	Transaction struct {
		ClientIP  string `json:"client_ip"`
		Timestamp string `json:"timestamp"`
		Request   struct {
			URI     string              `json:"uri"`
			Headers map[string][]string `json:"headers"`
		} `json:"request"`
	} `json:"transaction"`
	Messages []struct {
		Message string `json:"message"`
		Data    struct {
			ID     any    `json:"id"`
			File   string `json:"file"`
			Severity string `json:"severity"`
		} `json:"data"`
	} `json:"messages"`
}

// ParseCorazaAuditLine maps one JSON audit line to a ParsedHit.
// Unknown / non-JSON lines return an error (caller skips).
func ParseCorazaAuditLine(line string) (ParsedHit, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return ParsedHit{}, fmt.Errorf("empty line")
	}
	var raw corazaAuditLine
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return ParsedHit{}, err
	}
	host := firstHeader(raw.Transaction.Request.Headers, "Host", "host")
	host = strings.TrimSpace(strings.Split(host, ":")[0])
	path := strings.TrimSpace(raw.Transaction.Request.URI)
	if path == "" {
		path = "/"
	}
	rule := ""
	detail := ""
	if len(raw.Messages) > 0 {
		msg := raw.Messages[0]
		rule = fmt.Sprintf("%v", msg.Data.ID)
		if rule == "" || rule == "<nil>" {
			rule = ""
		}
		detail = strings.TrimSpace(msg.Message)
		if len(detail) > 512 {
			detail = detail[:512]
		}
	}
	ts := time.Now().UTC()
	if raw.Transaction.Timestamp != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, raw.Transaction.Timestamp); err == nil {
			ts = parsed.UTC()
		} else if parsed, err := time.Parse(time.RFC3339, raw.Transaction.Timestamp); err == nil {
			ts = parsed.UTC()
		}
	}
	action := "detect"
	for _, m := range raw.Messages {
		sev := strings.ToLower(m.Data.Severity)
		if strings.Contains(sev, "critical") || strings.Contains(strings.ToLower(m.Message), "denied") {
			action = "block"
			break
		}
	}
	return ParsedHit{
		TS:       ts,
		Hostname: host,
		Path:     path,
		Rule:     rule,
		Action:   action,
		IP:       strings.TrimSpace(raw.Transaction.ClientIP),
		Detail:   detail,
	}, nil
}

func firstHeader(h map[string][]string, keys ...string) string {
	for _, k := range keys {
		if vals, ok := h[k]; ok && len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}
