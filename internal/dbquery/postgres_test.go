package dbquery

import "testing"

func TestValidateReadOnlySQL(t *testing.T) {
	if err := validateReadOnlySQL("SELECT 1"); err != nil {
		t.Fatal(err)
	}
	if err := validateReadOnlySQL("WITH x AS (SELECT 1) SELECT * FROM x"); err != nil {
		t.Fatal(err)
	}
	if err := validateReadOnlySQL("DELETE FROM users"); err == nil {
		t.Fatal("expected reject")
	}
	if err := validateReadOnlySQL("SELECT 1; SELECT 2"); err == nil {
		t.Fatal("expected reject semicolon")
	}
}
