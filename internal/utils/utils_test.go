package utils

import (
	"testing"
)

func TestCalculateTolerancePercentage(t *testing.T) {
	tests := []struct {
		tolerance float64
		precision int
		expected  string
	}{
		{0.1234, 2, "12.34%"},
		{0.1234, 3, "12.340%"},
		{1.0, 0, "100%"},
		{0.5678, 1, "56.8%"},
		{0.1, 4, "10.0000%"},
		{0.9876, 5, "98.76000%"},
		{0.0, 2, "0.00%"},
		{0.9999, 2, "99.99%"},
		{0.0001, 2, "0.01%"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := CalculateTolerancePercentage(test.tolerance, test.precision)
			if result != test.expected {
				t.Errorf("CalculateTolerancePercentage(%v, %v) = %v; want %v", test.tolerance, test.precision, result, test.expected)
			}
		})
	}
}