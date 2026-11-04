package proxy

// staticFileHandles builds Caddy handlers for a static site root.
func staticFileHandles(r StaticRoute) []any {
	indexNames := r.IndexNames
	if len(indexNames) == 0 {
		indexNames = []string{"index.html"}
	}
	return []any{map[string]any{
		"handler":     "file_server",
		"root":        r.Root,
		"index_names": indexNames,
	}}
}
