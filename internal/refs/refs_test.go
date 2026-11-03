package refs

import "testing"

func TestResolvePostgresRef(t *testing.T) {
	catalog := Catalog{
		"db": {
			"DATABASE_URL": "postgres://u:p@db:5432/app?sslmode=disable",
			"PGHOST":       "db",
		},
	}
	got, err := Resolve("DATABASE_URL=${{db.DATABASE_URL}}", catalog)
	if err != nil {
		t.Fatal(err)
	}
	if got != "DATABASE_URL=postgres://u:p@db:5432/app?sslmode=disable" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveUnknownSlug(t *testing.T) {
	_, err := Resolve("${{missing.PGHOST}}", Catalog{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveCycle(t *testing.T) {
	catalog := Catalog{
		"a": {"X": "${{b.Y}}"},
		"b": {"Y": "${{a.X}}"},
	}
	_, err := Resolve("${{a.X}}", catalog)
	if err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestResolveNestedReplacement(t *testing.T) {
	catalog := Catalog{
		"db": {"PGPASSWORD": "secret"},
		"app": {"WRAPPER": "pw=${{db.PGPASSWORD}}"},
	}
	got, err := Resolve("VAL=${{app.WRAPPER}}", catalog)
	if err != nil {
		t.Fatal(err)
	}
	if got != "VAL=pw=secret" {
		t.Fatalf("got %q", got)
	}
}
