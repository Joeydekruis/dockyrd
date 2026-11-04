package proxy

import (
	"encoding/json"
	"sort"
)

// AppRoute is one HTTPS host → container upstream mapping.
type AppRoute struct {
	Hostname     string
	UpstreamDial string
}

// StaticRoute serves files from a host filesystem root (static sites).
type StaticRoute struct {
	Hostname string
	Root     string
}

// RenderInput controls global Caddy options.
type RenderInput struct {
	ACMEEmail     string
	PanelHost     string
	PanelUpstream string
	AppRoutes     []AppRoute
	StaticRoutes  []StaticRoute
	AdminListen   string
	AccessLog     string
}

// Render builds the full Caddy JSON document (briefing bijlage B).
func Render(in RenderInput) ([]byte, error) {
	if in.AdminListen == "" {
		in.AdminListen = "unix//run/caddy/admin.sock"
	}
	if in.AccessLog == "" {
		in.AccessLog = "/var/lib/dockyard/caddy/access.log"
	}
	if in.ACMEEmail == "" {
		in.ACMEEmail = "ops@localhost"
	}

	routes := make([]any, 0, len(in.AppRoutes)+len(in.StaticRoutes)+1)

	if in.PanelHost != "" && in.PanelUpstream != "" {
		routes = append(routes, map[string]any{
			"match": []any{map[string]any{"host": []string{in.PanelHost}}},
			"handle": []any{map[string]any{
				"handler": "reverse_proxy",
				"upstreams": []any{map[string]any{"dial": in.PanelUpstream}},
			}},
			"terminal": true,
		})
	}

	sortedApps := append([]AppRoute(nil), in.AppRoutes...)
	sort.Slice(sortedApps, func(i, j int) bool {
		return sortedApps[i].Hostname < sortedApps[j].Hostname
	})
	for _, r := range sortedApps {
		routes = append(routes, map[string]any{
			"match": []any{map[string]any{"host": []string{r.Hostname}}},
			"handle": []any{map[string]any{
				"handler": "reverse_proxy",
				"upstreams": []any{map[string]any{"dial": r.UpstreamDial}},
				"load_balancing": map[string]any{"try_duration": "5s"},
			}},
			"terminal": true,
		})
	}

	sortedStatic := append([]StaticRoute(nil), in.StaticRoutes...)
	sort.Slice(sortedStatic, func(i, j int) bool {
		return sortedStatic[i].Hostname < sortedStatic[j].Hostname
	})
	for _, r := range sortedStatic {
		routes = append(routes, map[string]any{
			"match": []any{map[string]any{"host": []string{r.Hostname}}},
			"handle": []any{map[string]any{
				"handler": "file_server",
				"root":    r.Root,
			}},
			"terminal": true,
		})
	}

	doc := map[string]any{
		"storage": map[string]any{
			"module": "file_system",
			"root":   "/var/lib/dockyard/caddy/certificates",
		},
		"admin": map[string]any{"listen": in.AdminListen},
		"logging": map[string]any{
			"logs": map[string]any{
				"access": map[string]any{
					"writer": map[string]any{
						"output":       "file",
						"filename":     in.AccessLog,
						"roll_size_mb": 50,
						"roll_keep":    3,
					},
					"encoder": map[string]any{"format": "json"},
					"include": []string{"http.log.access.dy"},
				},
			},
		},
		"apps": map[string]any{
			"http": map[string]any{
				"servers": map[string]any{
					"dy": map[string]any{
						"listen": []string{":443", ":80"},
						"logs":   map[string]any{"default_logger_name": "dy"},
						"routes": routes,
					},
				},
			},
			"tls": map[string]any{
				"automation": map[string]any{
					"policies": []any{
						map[string]any{
							"issuers": []any{
								map[string]any{
									"module": "acme",
									"email":  in.ACMEEmail,
								},
							},
						},
					},
				},
			},
		},
	}
	return json.MarshalIndent(doc, "", "  ")
}
