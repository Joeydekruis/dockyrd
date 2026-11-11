package hostops

import "testing"

func TestParseSSHDPasswordAuth(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cfg  string
		want bool
	}{
		{"passwordauthentication no\n", false},
		{"passwordauthentication yes\n", true},
		{"match user foo\npasswordauthentication yes\npasswordauthentication no\n", false},
	}
	for _, tc := range cases {
		if got := parseSSHDPasswordAuth(tc.cfg); got != tc.want {
			t.Fatalf("cfg=%q got=%v want=%v", tc.cfg, got, tc.want)
		}
	}
}
