package utils

import (
	"file-finder/internal/types"
	"os"
	"path/filepath"
	"testing"
)

func TestClearConsole(t *testing.T) {
	// This function clears the console, so it's difficult to unit test.
	// You can test if the function runs without errors.
	ClearConsole()
}

func TestFileCount(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "sub")
	os.Mkdir(subDir, 0755)
	file := filepath.Join(subDir, "testfile.mp4")
	os.WriteFile(file, []byte("test"), 0644)

	// Test FileCount function
	count, err := FileCount(tempDir, types.Video)
	if err != nil {
		t.Fatalf("FileCount() returned error: %v", err)
	}
	if count != 1 {
		t.Errorf("FileCount() returned %d, want %d", count, 1)
	}
}

func TestGetOperatorSizeMatches(t *testing.T) {
	tests := []struct {
		operator types.OperatorType
		fileSize int64
		infoSize int64
		expected bool
	}{
		{types.EqualToType, 100, 100, true},
		{types.LessThanType, 100, 50, true},
		{types.LessThanEqualToType, 100, 100, true},
		{types.GreaterThanType, 100, 150, true},
		{types.GreaterThanEqualToType, 100, 100, true},
		{types.GreaterThanEqualToType, 100, 50, false},
	}

	for _, tt := range tests {
		result := GetOperatorSizeMatches(tt.operator, tt.fileSize, tt.infoSize)
		if result != tt.expected {
			t.Errorf("GetOperatorSizeMatches(%v, %d, %d) = %v; want %v", tt.operator, tt.fileSize, tt.infoSize, result, tt.expected)
		}
	}
}

func TestGetOperatorToString(t *testing.T) {
	tests := []struct {
		operator types.OperatorType
		expected string
	}{
		{types.EqualToType, "equal to"},
		{types.LessThanType, "less than"},
		{types.LessThanEqualToType, "less than or equal to"},
		{types.GreaterThanType, "greater than"},
		{types.GreaterThanEqualToType, "greater than or equal to"},
	}

	for _, tt := range tests {
		result := GetOperatorToString(tt.operator)
		if result != tt.expected {
			t.Errorf("GetOperatorToString(%v) = %q; want %q", tt.operator, result, tt.expected)
		}
	}
}

func TestGetFileTypeToString(t *testing.T) {
	tests := []struct {
		fileType types.FileType
		expected string
	}{
		{types.Any, "Any"},
		{types.Video, "Video"},
		{types.Image, "Image"},
		{types.Archive, "Archive"},
		{types.Documents, "Documents"},
	}

	for _, tt := range tests {
		result := GetFileTypeToString(tt.fileType)
		if result != tt.expected {
			t.Errorf("GetFileTypeToString(%v) = %q; want %q", tt.fileType, result, tt.expected)
		}
	}
}

func TestIsFileOfType(t *testing.T) {
	tests := []struct {
		ext      string
		fileType types.FileType
		expected bool
	}{
		{".mp4", types.Video, true},
		{".jpg", types.Image, true},
		{".zip", types.Archive, true},
		{".pdf", types.Documents, true},
		{".exe", types.Video, false},
		{".unknown", types.Any, false},
	}

	for _, tt := range tests {
		result := IsFileOfType(tt.ext, tt.fileType)
		if result != tt.expected {
			t.Errorf("IsFileOfType(%q, %v) = %v; want %v", tt.ext, tt.fileType, result, tt.expected)
		}
	}
}

func TestFormatPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"C:\\Users\\User\\file.txt", "C:/Users/User/file.txt"}, // Test for Windows
		{"/home/user/file.txt", "/home/user/file.txt"},          // Test for Unix-like
	}

	for _, tt := range tests {
		result := formatPath(tt.path)
		if result != tt.expected {
			t.Errorf("formatPath(%q) = %q; want %q", tt.path, result, tt.expected)
		}
	}
}

func TestRenderResultsTable(t *testing.T) {
	// This function is difficult to unit test because it involves output.
	// You can test if the function runs without errors.
	results := []types.DirectoryResult{
		{Directory: "dir1", Count: 10},
		{Directory: "dir2", Count: 20},
	}
	totalCount := 30

	RenderResultsTable(results, totalCount)
}

func TestConvertSizeToBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"1 B", 1, false},
		{"10 B", 10, false},
		{"100 B", 100, false},
		{"1000 B", 1000, false},
		{"1 KB", 1024, false},
		{"10 KB", 10240, false},
		{"100 KB", 102400, false},
		{"1000 KB", 1024000, false},
		{"1 MB", 1024 * 1024, false},
		{"10 MB", 10 * 1024 * 1024, false},
		{"100 MB", 100 * 1024 * 1024, false},
		{"1000 MB", 1000 * 1024 * 1024, false},
		{"1 GB", 1024 * 1024 * 1024, false},
		{"10 GB", 10 * 1024 * 1024 * 1024, false},
		{"100 GB", 100 * 1024 * 1024 * 1024, false},
		{"1000 GB", 1000 * 1024 * 1024 * 1024, false},
		{"1 PB", 1024 * 1024 * 1024 * 1024, false},
		{"10 PB", 10 * 1024 * 1024 * 1024 * 1024, false},
		{"100 PB", 100 * 1024 * 1024 * 1024 * 1024, false},
		{"1000 PB", 1000 * 1024 * 1024 * 1024 * 1024, false},
		// Invalid cases
		{"", 0, true},
		{"1 ZB", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		result, err := ConvertSizeToBytes(test.input)
		if (err != nil) != test.hasError {
			t.Errorf("ConvertSizeToBytes(%s) unexpected error: %v", test.input, err)
		}
		if result != test.expected {
			t.Errorf("ConvertSizeToBytes(%s) = %d; want %d", test.input, result, test.expected)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{1, "1.00 B"},
		{10, "10.00 B"},
		{100, "100.00 B"},
		{1000, "1000.00 B"},
		{1024, "1.00 KB"},
		{10240, "10.00 KB"},
		{102400, "100.00 KB"},
		{1024000, "1000.00 KB"},
		{1024 * 1024, "1.00 MB"},
		{10 * 1024 * 1024, "10.00 MB"},
		{100 * 1024 * 1024, "100.00 MB"},
		{1000 * 1024 * 1024, "1000.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{10 * 1024 * 1024 * 1024, "10.00 GB"},
		{100 * 1024 * 1024 * 1024, "100.00 GB"},
		{1000 * 1024 * 1024 * 1024, "1000.00 GB"},
		{1024 * 1024 * 1024 * 1024, "1.00 PB"},
		{10 * 1024 * 1024 * 1024 * 1024, "10.00 PB"},
		{100 * 1024 * 1024 * 1024 * 1024, "100.00 PB"},
		{1000 * 1024 * 1024 * 1024 * 1024, "1000.00 PB"},
	}

	for _, test := range tests {
		result := FormatSize(test.input)
		if result != test.expected {
			t.Errorf("FormatSize(%d) = %s; want %s", test.input, result, test.expected)
		}
	}
}
