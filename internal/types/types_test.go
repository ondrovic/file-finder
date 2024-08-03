package types

import (
	"testing"
)

func TestFileTypeValues(t *testing.T) {
	// Test if the FileType constants are correctly set
	tests := []struct {
		name     string
		fileType FileType
		expected int
	}{
		{"Any", Any, 0},
		{"Video", Video, 1},
		{"Image", Image, 2},
		{"Archive", Archive, 3},
		{"Documents", Documents, 4},
	}

	for _, tt := range tests {
		if int(tt.fileType) != tt.expected {
			t.Errorf("FileType %s expected %d, got %d", tt.name, tt.expected, tt.fileType)
		}
	}
}

func TestOperatorTypeValues(t *testing.T) {
	// Test if the OperatorType constants are correctly set
	tests := []struct {
		name         string
		operatorType OperatorType
		expected     int
	}{
		{"EqualTo", EqualToType, 0},
		{"GreaterThan", GreaterThanType, 1},
		{"GreaterThanEqualTo", GreaterThanEqualToType, 2},
		{"LessThan", LessThanType, 3},
		{"LessThanEqualTo", LessThanEqualToType, 4},
	}

	for _, tt := range tests {
		if int(tt.operatorType) != tt.expected {
			t.Errorf("OperatorType %s expected %d, got %d", tt.name, tt.expected, tt.operatorType)
		}
	}
}

func TestNewVideoFinder(t *testing.T) {
	// Test if the NewVideoFinder function returns a valid VideoFinder instance
	vf := NewVideoFinder()

	if vf == nil {
		t.Fatal("NewVideoFinder() returned nil")
	}

	if len(vf.Results) != 0 {
		t.Errorf("NewVideoFinder().Results should be empty, got %d", len(vf.Results))
	}
}

func TestFileExtensions(t *testing.T) {
	// Test if the FileExtensions map contains the correct file extensions for each FileType
	tests := []struct {
		fileType   FileType
		extensions []string
	}{
		{Video, []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpg", ".mpeg", ".ts"}},
		{Image, []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".svg", ".raw", ".heic", ".ico"}},
		{Archive, []string{".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz", ".iso", ".tgz", ".tbz2"}},
		{Documents, []string{".docx", ".doc", ".pdf", ".txt", ".rtf", ".odt", ".xlsx", ".xls", ".pptx", ".ppt", ".csv", ".md", ".pages"}},
	}

	for _, tt := range tests {
		for _, ext := range tt.extensions {
			if !FileExtensions[tt.fileType][ext] {
				t.Errorf("FileExtensions[%d] should contain %s", tt.fileType, ext)
			}
		}
	}
}
