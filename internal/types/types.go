package types

import (
	commonTypes "github.com/ondrovic/common/types"
)

// FileFinder struct remains the same
type FileFinder struct {
	RootDir          string
	DeleteFlag       bool
	DetailedListFlag bool
	FileSize         string
	FileType         commonTypes.FileType
	OperatorType     commonTypes.OperatorType
	Tolerance        float64
	Results          map[string][]string
}

// DirectoryResults struct for the results
type DirectoryResult struct {
	Directory string
	Count     int
}

// EntryResults struct for more in depth entry info
type EntryResults struct {
	Directory string
	FileName  string
	FileSize  string
}

//NewFileFinder initializes a new FileFinder object
func NewFileFinder() *FileFinder {
	return &FileFinder{
		Results: make(map[string][]string),
	}
}
