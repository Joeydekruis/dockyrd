package refs

// RefSlug names a referenced service slug inside a placeholder.
type RefSlug struct {
	Slug string
	Key  string
}

// ListRefs returns every ${{slug.VAR}} reference in s (deduped, stable order).
func ListRefs(s string) []RefSlug {
	if s == "" || !ContainsRef(s) {
		return nil
	}
	seen := make(map[RefSlug]struct{})
	var out []RefSlug
	for _, m := range refPattern.FindAllStringSubmatch(s, -1) {
		if len(m) < 3 {
			continue
		}
		r := RefSlug{Slug: m[1], Key: m[2]}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out
}

// RefSlugsUnique returns distinct slugs referenced in s.
func RefSlugsUnique(s string) []string {
	refs := ListRefs(s)
	if len(refs) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var out []string
	for _, r := range refs {
		if _, ok := seen[r.Slug]; ok {
			continue
		}
		seen[r.Slug] = struct{}{}
		out = append(out, r.Slug)
	}
	return out
}
