package metrics

import "testing"

func TestNormalizeHostCPUPercentLegacy(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   float64
		want float64
	}{
		{0.42, 0.42},
		{42, 0.42},
		{100, 1},
		{0, 0},
	}
	for _, tc := range cases {
		got := NormalizeHostCPUFraction(tc.in)
		if got != tc.want {
			t.Fatalf("NormalizeHostCPUFraction(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
