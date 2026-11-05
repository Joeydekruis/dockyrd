package refs

import "testing"

func TestListRefs_dedupes(t *testing.T) {
	s := "A=${{db.DATABASE_URL}} B=${{db.DATABASE_URL}} C=${{cache.REDIS_URL}}"
	got := ListRefs(s)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
}

func TestRefSlugsUnique(t *testing.T) {
	slugs := RefSlugsUnique("${{db.X}} ${{cache.Y}} ${{db.Z}}")
	if len(slugs) != 2 || slugs[0] != "db" || slugs[1] != "cache" {
		t.Fatalf("%v", slugs)
	}
}
