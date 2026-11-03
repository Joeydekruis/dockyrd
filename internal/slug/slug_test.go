package slug

import "testing"

func TestFromName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Client X", "client-x"},
		{"  Hello World!  ", "hello-world"},
		{"---", "project"},
		{"API v2", "api-v2"},
	}
	for _, tc := range tests {
		if got := FromName(tc.in); got != tc.want {
			t.Fatalf("FromName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
