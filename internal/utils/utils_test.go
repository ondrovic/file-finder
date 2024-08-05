package utils

import (
	"errors"
	"file-finder/internal/types"
	"testing"
)

// #region Public Functions Tests
func TestClearConsole(t *testing.T) {
	// This function clears the console, so it's difficult to unit test.
	// You can test if the function runs without errors.
	ClearConsole()
}

func TestToFileType(t *testing.T) {
	tests := []struct {
		input    string
		expected types.FileType
	}{
		{"any", types.Any},
		{"video", types.Video},
		{"image", types.Image},
		{"archive", types.Archive},
		{"documents", types.Documents},
		{"ANY", types.Any},        // Case-insensitive check
		{"ViDeO", types.Video},    // Mixed case check
		{"invalid", ""},           // Invalid input check
		{"", ""},                  // Empty string check
	}

	for _, test := range tests {
		result := ToFileType(test.input)
		if result != test.expected {
			t.Errorf("ToFileType(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestToOperatorType(t *testing.T) {
	tests := []struct {
		input    string
		expected types.OperatorType
	}{
		{"equal to", types.EqualTo},
		{"equalto", types.EqualTo},
		{"equal", types.EqualTo},
		{"==", types.EqualTo},
		{"greater than", types.GreaterThan},
		{"greaterthan", types.GreaterThan},
		{">", types.GreaterThan},
		{"greater than or equal to", types.GreaterThanEqualTo},
		{"greaterthanorequalto", types.GreaterThanEqualTo},
		{">=", types.GreaterThanEqualTo},
		{"less than", types.LessThan},
		{"lessthan", types.LessThan},
		{"<", types.LessThan},
		{"less than or equal to", types.LessThanEqualTo},
		{"lessthanorequalto", types.LessThanEqualTo},
		{"<=", types.LessThanEqualTo},
		{"INVALID", ""}, // Invalid input case
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := ToOperatorType(test.input)
			if result != test.expected {
				t.Errorf("ToOperatorType(%q) = %v; want %v", test.input, result, test.expected)
			}
		})
	}
}

func TestConvertSizeToBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		err      error
	}{
		{"1 B", 1, nil},
		{"10 KB", 10 * 1024, nil},
		{"1 MB", 1 * 1024 * 1024, nil},
		{"5 GB", 5 * 1024 * 1024 * 1024, nil},
		{"100 GB", 100 * 1024 * 1024 * 1024, nil},
		{"2.5 TB", 2.5 * 1024 * 1024 * 1024 * 1024, nil},
		{"1 kB", 1 * 1024, nil}, // Check case insensitivity
		{"1000M", 0, errors.New("invalid size unit")}, // No matching unit
		{"", 0, errors.New("size cannot be empty")},
		{"1000", 0, errors.New("invalid size format")},
		{"1000 XYZ", 0, errors.New("invalid size unit")},
		{"not a size", 0, errors.New("invalid size format")},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := ConvertSizeToBytes(test.input)

			if result != test.expected {
				t.Errorf("ConvertSizeToBytes(%q) = %v; want %v", test.input, result, test.expected)
			}

			if (err != nil && test.err == nil) || (err == nil && test.err != nil) || (err != nil && test.err != nil && err.Error() != test.err.Error()) {
				t.Errorf("ConvertSizeToBytes(%q) error = %v; want %v", test.input, err, test.err)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1, "1.00 B"},
		{1024, "1.00 KB"},            // 1024 bytes exactly 1 KB
		{2048, "2.00 KB"},            // 2048 bytes should be 2 KB
		{1048576, "1.00 MB"},         // 1048576 bytes exactly 1 MB
		{2097152, "2.00 MB"},         // 2097152 bytes should be 2 MB
		{1073741824, "1.00 GB"},      // 1073741824 bytes exactly 1 GB
		{2147483648, "2.00 GB"},      // 2147483648 bytes should be 2 GB
		{1125899906842624, "1.00 PB"},// 1125899906842624 bytes exactly 1 PB
		{2251799813685248, "2.00 PB"},// 2251799813685248 bytes should be 2 PB
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := FormatSize(test.bytes)
			if result != test.expected {
				t.Errorf("FormatSize(%v) = %v; want %v", test.bytes, result, test.expected)
			}
		})
	}
}

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

func TestCalculateToleranceBytes(t *testing.T) {
	tests := []struct {
		sizeStr   string
		tolerance float64
		expected  int64
		wantErr   bool
	}{
		{"1 KB", 10, 1126, false},             // 1 KB + 10% tolerance
		{"1 MB", 50, 1572864, false},         // 1 MB + 50% tolerance
		{"100 B", 100, 200, false},           // 100 B + 100% tolerance
		{"10 GB", 0, 10737418240, false},    // 10 GB + 0% tolerance
		{"2.5 KB", 20, 3072, false},          // 2.5 KB + 20% tolerance
		{"10 MB", 25, 13107200, false},       // 10 MB + 25% tolerance
		{"1000 B", 0, 1000, false},           // 1000 B + 0% tolerance
		{"5 GB", -10, 4831838208, false},     // 5 GB - 10% tolerance
		{"500 KB", 200, 1536000, false},      // 500 KB + 200% tolerance
		{"", 10, 0, true},                    // Empty size string should return an error
		{"1 KB", -10, 921, false},            // 1 KB - 10% tolerance
	}

	for _, test := range tests {
		t.Run(test.sizeStr, func(t *testing.T) {
			result, err := CalculateToleranceBytes(test.sizeStr, test.tolerance)
			if (err != nil) != test.wantErr {
				t.Errorf("CalculateToleranceBytes(%v, %v) error = %v; wantErr %v", test.sizeStr, test.tolerance, err, test.wantErr)
				return
			}
			if result != test.expected {
				t.Errorf("CalculateToleranceBytes(%v, %v) = %v; want %v", test.sizeStr, test.tolerance, result, test.expected)
			}
		})
	}
}

// #endregion