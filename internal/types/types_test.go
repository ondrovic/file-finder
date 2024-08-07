package types

import (
	"testing"
)

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