package links

import "testing"

func TestValidatePair_appToPostgres(t *testing.T) {
	if err := ValidatePair("app", "postgres"); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePair_rejectsStatic(t *testing.T) {
	if err := ValidatePair("app", "static_site"); err == nil {
		t.Fatal("expected error")
	}
}
