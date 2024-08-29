package types

import (
	sharedTypes "github.com/ondrovic/common/types"
)

type FileInfo struct {
	Path string
	Size int64
	// Hash string
}

// REMOVE: once refactored
// // FileFinder struct remains the same
// type FileFinder struct {
// 	DisplayApplicationBanner bool
// 	DisplayDetailedResults   bool
// 	FileNameFilter           string
// 	FileSizeFilter           string
// 	FileTypeFilter           sharedTypes.FileType
// 	ListDuplicateFiles       bool
// 	OperatorTypeFilter       sharedTypes.OperatorType
// 	RemoveFiles              bool
// 	Results                  map[string][]string
// 	RootDirectory            string
// 	ToleranceSize            float64
// }

// DirectoryResults struct for the results
type DirectoryResult struct {
	Directory string
	Count     int
}

// EntryResult struct for more in depth entry info
type EntryResult struct {
	Directory string
	FileName  string
	FileSize  string
}

// NewCliFlags initializes a new FileFinder object
func NewCliFlags() *CliFlags {
	return &CliFlags{
		Results: make(map[string][]string),
	}
}

type CliFlags struct {
	DisplayApplicationBanner bool
	DisplayDetailedResults   bool
	FileNameFilter           string
	FileSizeFilter           string
	FileTypeFilter           sharedTypes.FileType
	// ListDuplicateFiles       bool //TODO: need to work on this
	OperatorTypeFilter sharedTypes.OperatorType
	RemoveFiles        bool
	Results            map[string][]string
	RootDirectory      string
	ToleranceSize      float64
	// TODO: SortDescending
}
