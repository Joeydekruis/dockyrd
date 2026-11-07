package proxy

// staticFileHandles builds Caddy handlers for a static site root.
func staticFileHandles(r StaticRoute) []any {
	indexNames := r.IndexNames
	if len(indexNames) == 0 {
		indexNames = []string{"index.html"}
	}
	fs := map[string]any{
		"handler":     "file_server",
		"root":        r.Root,
		"index_names": indexNames,
	}
	tail := []any{
		staticImmutableCacheHandler(),
		staticHTMLNoCacheHandler(),
		fs,
	}
	if r.Security != nil {
		return prependSecurity(*r.Security, tail)
	}
	return prependBaseline(tail)
}
