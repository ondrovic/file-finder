package types

import (
	commonTypes "github.com/ondrovic/common/types"
)

type FileInfo struct {
	Path string
	Size int64
	Hash string
}

// FileFinder struct remains the same
type FileFinder struct {
	DisplayApplicationBanner bool
	DisplayDetailedResults   bool
	FileNameFilter           string
	FileSizeFilter           string
	FileTypeFilter           commonTypes.FileType
	ListDuplicateFiles       bool
	OperatorTypeFilter       commonTypes.OperatorType
	RemoveFiles              bool
	Results                  map[string][]string
	RootDirectory            string
	ToleranceSize            float64
}

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

// NewFileFinder initializes a new FileFinder object
func NewFileFinder() *FileFinder {
	return &FileFinder{
		Results: make(map[string][]string),
	}
}
