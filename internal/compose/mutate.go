package compose

import (
	"fmt"
	"strings"
)

const appsCgroupParent = "dy-apps.slice"

// MutateInput controls dockyard-specific compose mutations.
type MutateInput struct {
	ServiceID      string
	DeploymentID   string
	ServerID       string
	ProjectID      string
	StackDataDir   string
	AllowPublished map[string]struct{}
}

// Mutate applies labels, restart, logging, cgroup, aliases, and strips host ports.
func Mutate(doc map[string]any, in MutateInput) (map[string]any, []Issue) {
	if doc == nil {
		return nil, []Issue{{Rule: "document", Message: "empty compose document"}}
	}
	out := cloneMap(doc)
	services, _ := out["services"].(map[string]any)
	if len(services) == 0 {
		return out, nil
	}
	opts := PolicyOptions{
		StackDataDir:          in.StackDataDir,
		PublishedPortsAllowed: in.AllowPublished,
	}
	var warnings []Issue
	newServices := make(map[string]any, len(services))
	for name, raw := range services {
		svc, ok := raw.(map[string]any)
		if !ok {
			newServices[name] = raw
			continue
		}
		mutated, w := mutateService(name, cloneMap(svc), in)
		warnings = append(warnings, w...)
		warnings = append(warnings, CheckPolicy(map[string]any{"services": map[string]any{name: mutated}}, opts)...)
		newServices[name] = mutated
	}
	out["services"] = newServices
	out["networks"] = ensureDefaultNetwork(out["networks"], in)
	return out, warnings
}

func mutateService(name string, svc map[string]any, in MutateInput) (map[string]any, []Issue) {
	var warnings []Issue
	allowed := in.AllowPublished
	if allowed == nil {
		allowed = map[string]struct{}{}
	}
	if ports, ok := svc["ports"].([]any); ok && len(ports) > 0 {
		if _, ok := allowed[name]; !ok {
			delete(svc, "ports")
			warnings = append(warnings, Issue{
				Rule:    "ports",
				Warning: true,
				Message: fmt.Sprintf("service %q: stripped host ports (use a Dockyard domain)", name),
			})
		}
	}
	labels, _ := svc["labels"].(map[string]any)
	if labels == nil {
		labels = map[string]any{}
	}
	labels["dev.dockyard.managed"] = "true"
	labels["dev.dockyard.server"] = in.ServerID
	labels["dev.dockyard.project"] = in.ProjectID
	labels["dev.dockyard.service"] = in.ServiceID
	labels["dev.dockyard.kind"] = "compose"
	labels["dev.dockyard.deployment"] = in.DeploymentID
	labels["com.docker.compose.service"] = name
	labels["com.docker.compose.project"] = ProjectName(in.ServiceID)
	svc["labels"] = labels

	if _, ok := svc["restart"]; !ok {
		svc["restart"] = "unless-stopped"
	}
	svc["logging"] = map[string]any{
		"driver": "json-file",
		"options": map[string]any{
			"max-size": "10m",
			"max-file": "3",
		},
	}
	svc["cgroup_parent"] = appsCgroupParent
	if _, ok := svc["security_opt"]; !ok {
		svc["security_opt"] = []any{"no-new-privileges:true"}
	}
	alias := ServiceAlias(in.ServiceID, name)
	networks := map[string]any{}
	if raw, ok := svc["networks"]; ok {
		switch t := raw.(type) {
		case []any:
			for _, n := range t {
				key, _ := n.(string)
				if key != "" {
					networks[key] = map[string]any{"aliases": []any{alias}}
				}
			}
		case map[string]any:
			for key := range t {
				networks[key] = map[string]any{"aliases": []any{alias}}
			}
		}
	}
	if len(networks) == 0 {
		networks["default"] = map[string]any{"aliases": []any{alias}}
	}
	svc["networks"] = networks
	return svc, warnings
}

func ensureDefaultNetwork(raw any, in MutateInput) map[string]any {
	netName := StackNetworkName(ProjectName(in.ServiceID))
	nets, _ := raw.(map[string]any)
	if nets == nil {
		nets = map[string]any{}
	}
	if _, ok := nets["default"]; !ok {
		nets["default"] = map[string]any{"name": netName}
	} else if def, ok := nets["default"].(map[string]any); ok {
		def["name"] = netName
		nets["default"] = def
	}
	return nets
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// ListServiceNames returns compose service keys from a document.
func ListServiceNames(doc map[string]any) []string {
	services, _ := doc["services"].(map[string]any)
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	return names
}

// ExposablePorts heuristically lists container ports for domain wiring.
func ExposablePorts(doc map[string]any) []Exposure {
	services, _ := doc["services"].(map[string]any)
	var out []Exposure
	for name, raw := range services {
		svc, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		found := false
		if expose, ok := svc["expose"].([]any); ok {
			for _, p := range expose {
				if port := parsePort(p); port > 0 {
					out = append(out, Exposure{Service: name, Port: port})
					found = true
				}
			}
		}
		if ports, ok := svc["ports"].([]any); ok {
			for _, p := range ports {
				if port := parsePortString(p); port > 0 {
					out = append(out, Exposure{Service: name, Port: port})
					found = true
				}
			}
		}
		if !found {
			out = append(out, Exposure{Service: name, Port: 80})
		}
	}
	return out
}

func parsePort(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case string:
		return parsePortString(t)
	default:
		return 0
	}
}

func parsePortString(v any) int {
	s, ok := v.(string)
	if !ok {
		return parsePort(v)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	parts := strings.Split(s, ":")
	last := parts[len(parts)-1]
	last = strings.TrimSuffix(last, "/tcp")
	last = strings.TrimSuffix(last, "/udp")
	var port int
	_, _ = fmt.Sscanf(last, "%d", &port)
	return port
}
