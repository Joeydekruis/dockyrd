package instance

import "testing"

func TestPanelHostname(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://dockyrd.derisory.nl", "dockyrd.derisory.nl"},
		{"https://dockyrd.derisory.nl/", "dockyrd.derisory.nl"},
		{"http://panel.example.com:8000", "panel.example.com"},
		{"panel.example.com", "panel.example.com"},
		{"", ""},
		{"https://127.0.0.1", ""},
	}
	for _, tc := range tests {
		if got := PanelHostname(tc.in); got != tc.want {
			t.Errorf("PanelHostname(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
