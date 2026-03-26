package oracle

import (
	"testing"
)

func TestIsDifferenceSignificant(t *testing.T) {
	tests := []struct {
		name   string
		a      uint64
		b      uint64
		sig    float64
		expect bool
	}{
		{name: "test 1", a: 1, b: 1, sig: 0.05, expect: false},
		{name: "test 2", a: 4, b: 1, sig: 0.25, expect: true},
		{name: "test 3", a: 3, b: 1, sig: 0.1, expect: true},
		{name: "test 4", a: 4, b: 1, sig: 0.9, expect: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isDifferenceSignificant(tc.a, tc.b, tc.sig)
			if result != tc.expect {
				t.Fatalf("mismatch %s", tc.name)
			}
		})
	}
}
