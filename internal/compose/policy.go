package compose

import (
	"fmt"
	"strings"
)

// Issue is a policy violation or warning.
type Issue struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
	Warning bool   `json:"warning,omitempty"`
}

// PolicyOptions controls strictness.
type PolicyOptions struct {
	AllowUnsafe           bool
	StackDataDir          string
	PublishedPortsAllowed map[string]struct{}
}

// CheckPolicy validates normalized compose JSON (docker compose config --format json).
func CheckPolicy(doc map[string]any, opts PolicyOptions) []Issue {
	if doc == nil {
		return []Issue{{Rule: "document", Message: "empty compose document"}}
	}
	services, _ := doc["services"].(map[string]any)
	if len(services) == 0 {
		return []Issue{{Rule: "services", Message: "at least one service required"}}
	}
	var issues []Issue
	for name, raw := range services {
		svc, ok := raw.(map[string]any)
		if !ok {
			issues = append(issues, Issue{Rule: "services", Message: fmt.Sprintf("service %q is invalid", name)})
			continue
		}
		issues = append(issues, checkServicePolicy(name, svc, opts)...)
	}
	return issues
}

func checkServicePolicy(name string, svc map[string]any, opts PolicyOptions) []Issue {
	var issues []Issue
	if opts.AllowUnsafe {
		return issues
	}
	if truthy(svc["privileged"]) {
		issues = append(issues, Issue{Rule: "privileged", Message: fmt.Sprintf("service %q: privileged is not allowed", name)})
	}
	for _, field := range []string{"network_mode", "pid", "ipc", "userns_mode"} {
		if isHostMode(svc[field]) {
			issues = append(issues, Issue{Rule: field, Message: fmt.Sprintf("service %q: %s host mode is not allowed", name, field)})
		}
	}
	if _, ok := svc["build"]; ok {
		issues = append(issues, Issue{Rule: "build", Message: fmt.Sprintf("service %q: build is not supported in v1 (use image:)", name)})
	}
	if cn, ok := svc["container_name"].(string); ok && strings.TrimSpace(cn) != "" {
		issues = append(issues, Issue{Rule: "container_name", Message: fmt.Sprintf("service %q: container_name is managed by dockyrd", name)})
	}
	issues = append(issues, checkMounts(name, svc, opts)...)
	issues = append(issues, checkCaps(name, svc)...)
	issues = append(issues, checkSecurityOpts(name, svc)...)
	issues = append(issues, checkDevices(name, svc)...)
	issues = append(issues, checkSysctls(name, svc)...)
	issues = append(issues, checkPorts(name, svc, opts)...)
	return issues
}

func checkPorts(name string, svc map[string]any, opts PolicyOptions) []Issue {
	ports, ok := svc["ports"].([]any)
	if !ok || len(ports) == 0 {
		return nil
	}
	allowed := opts.PublishedPortsAllowed
	if allowed == nil {
		allowed = map[string]struct{}{}
	}
	if _, ok := allowed[name]; ok {
		return nil
	}
	return []Issue{{
		Rule:    "ports",
		Warning: true,
		Message: fmt.Sprintf("service %q: host ports will be stripped (use a dockyrd domain)", name),
	}}
}

func checkMounts(name string, svc map[string]any, opts PolicyOptions) []Issue {
	volumes, ok := svc["volumes"].([]any)
	if !ok {
		return nil
	}
	prefix := strings.TrimSuffix(strings.TrimSpace(opts.StackDataDir), "/")
	if prefix != "" {
		prefix += "/"
	}
	var issues []Issue
	for _, raw := range volumes {
		src := volumeSource(raw)
		if src == "" {
			continue
		}
		if strings.Contains(src, "docker.sock") {
			issues = append(issues, Issue{Rule: "volumes", Message: fmt.Sprintf("service %q: mounting docker.sock is not allowed", name)})
			continue
		}
		if strings.HasPrefix(src, "/") {
			if prefix == "" || !strings.HasPrefix(src, prefix) {
				issues = append(issues, Issue{Rule: "volumes", Message: fmt.Sprintf("service %q: bind mount %q must stay under the stack directory", name, src)})
			}
		}
	}
	return issues
}

func volumeSource(raw any) string {
	switch v := raw.(type) {
	case string:
		parts := strings.Split(v, ":")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	case map[string]any:
		if s, ok := v["source"].(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func checkCaps(name string, svc map[string]any) []Issue {
	caps, ok := svc["cap_add"].([]any)
	if !ok {
		return nil
	}
	bad := map[string]struct{}{"SYS_ADMIN": {}, "NET_ADMIN": {}, "ALL": {}}
	for _, c := range caps {
		s, _ := c.(string)
		s = strings.ToUpper(strings.TrimSpace(s))
		if _, blocked := bad[s]; blocked {
			return []Issue{{Rule: "cap_add", Message: fmt.Sprintf("service %q: cap_add %q is not allowed", name, s)}}
		}
	}
	return nil
}

func checkSecurityOpts(name string, svc map[string]any) []Issue {
	opts, ok := svc["security_opt"].([]any)
	if !ok {
		return nil
	}
	for _, o := range opts {
		s, _ := o.(string)
		lower := strings.ToLower(s)
		if strings.Contains(lower, "seccomp=unconfined") || strings.Contains(lower, "apparmor=unconfined") {
			return []Issue{{Rule: "security_opt", Message: fmt.Sprintf("service %q: disabling seccomp/apparmor is not allowed", name)}}
		}
	}
	return nil
}

func checkDevices(name string, svc map[string]any) []Issue {
	if devs, ok := svc["devices"].([]any); ok && len(devs) > 0 {
		return []Issue{{Rule: "devices", Message: fmt.Sprintf("service %q: devices are not allowed", name)}}
	}
	return nil
}

func checkSysctls(name string, svc map[string]any) []Issue {
	if raw, ok := svc["sysctls"]; ok && raw != nil {
		switch t := raw.(type) {
		case map[string]any:
			if len(t) > 0 {
				return []Issue{{Rule: "sysctls", Message: fmt.Sprintf("service %q: sysctls are not allowed", name)}}
			}
		case []any:
			if len(t) > 0 {
				return []Issue{{Rule: "sysctls", Message: fmt.Sprintf("service %q: sysctls are not allowed", name)}}
			}
		}
	}
	return nil
}

func isHostMode(v any) bool {
	s, _ := v.(string)
	return strings.EqualFold(strings.TrimSpace(s), "host")
}

func truthy(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(strings.TrimSpace(t), "true")
	default:
		return false
	}
}

// HasBlockingErrors returns true when any issue is not a warning.
func HasBlockingErrors(issues []Issue) bool {
	for _, i := range issues {
		if !i.Warning {
			return true
		}
	}
	return false
}
