package types

import (
	"testing"
)

// TestFileTypeConstants checks that FileType constants have the expected values
func TestFileTypeConstants(t *testing.T) {
	tests := []struct {
		expected FileType
		actual   FileType
	}{
		{Any, "Any"},
		{Video, "Video"},
		{Image, "Image"},
		{Archive, "Archive"},
		{Documents, "Documents"},
	}

	for _, test := range tests {
		if test.expected != test.actual {
			t.Errorf("expected %s, got %s", test.expected, test.actual)
		}
	}
}

// TestOperatorTypeConstants checks that OperatorType constants have the expected values
func TestOperatorTypeConstants(t *testing.T) {
	tests := []struct {
		expected OperatorType
		actual   OperatorType
	}{
		{EqualTo, "Equal To"},
		{GreaterThan, "Greater Than"},
		{GreaterThanEqualTo, "Greater Than Or Equal To"},
		{LessThan, "Less Than"},
		{LessThanEqualTo, "Less Than Or Equal To"},
	}

	for _, test := range tests {
		if test.expected != test.actual {
			t.Errorf("expected %s, got %s", test.expected, test.actual)
		}
	}
}

// TestFileExtensions checks that FileExtensions map contains the expected file types
func TestFileExtensions(t *testing.T) {
	expectedExtensions := map[FileType][]string{
		Any: {
			"*.*",
		},
		Video: {
			".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpg", ".mpeg", ".ts",
		},
		Image: {
			".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".svg", ".raw", ".heic", ".ico",
		},
		Archive: {
			".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz", ".iso", ".tgz", ".tbz2",
		},
		Documents: {
			".docx", ".doc", ".pdf", ".txt", ".rtf", ".odt", ".xlsx", ".xls", ".pptx", ".ppt", ".csv", ".md", ".pages",
		},
	}

	for fileType, extensions := range expectedExtensions {
		for _, ext := range extensions {
			if !FileExtensions[fileType][ext] {
				t.Errorf("expected file type %s to contain extension %s", fileType, ext)
			}
		}
	}
}

// TestNewFileFinder checks that NewFileFinder initializes a FileFinder with default values
func TestNewFileFinder(t *testing.T) {
	ff := NewFileFinder()

	if ff == nil {
		t.Fatal("expected NewFileFinder to return a non-nil pointer")
	}

	if len(ff.Results) != 0 {
		t.Errorf("expected Results to be an empty map, got %v", ff.Results)
	}
}