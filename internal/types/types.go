package types

// FileType represents different types of files
type FileType string

// OperatorType represents different file size operations
type OperatorType string

// FileType const
const (
	Any       FileType = "Any"
	Video     FileType = "Video"
	Image     FileType = "Image"
	Archive   FileType = "Archive"
	Documents FileType = "Documents"
)

// OperatorType const
const (
	EqualTo            OperatorType = "Equal To"
	GreaterThan        OperatorType = "Greater Than"
	GreaterThanEqualTo OperatorType = "Greater Than Or Equal To"
	LessThan           OperatorType = "Less Than"
	LessThanEqualTo    OperatorType = "Less Than Or Equal To"
)

// FileFinder struct remains the same
type FileFinder struct {
	RootDir          string
	DeleteFlag       bool
	DetailedListFlag bool
	FileSize         string
	FileType         FileType
	OperatorType     OperatorType
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

// Unit struct for units
type Unit struct {
	Label string
	Size  int64
}

var (
	
	Units = []Unit{
		{"PB", 1 << 50}, // Petabyte
		{"TB", 1 << 40}, // Terabyte
		{"GB", 1 << 30}, // Gigabyte
		{"MB", 1 << 20}, // Megabyte
		{"KB", 1 << 10}, // Kilobyte
		{"B", 1},        // Byte
	}

	FileExtensions = map[FileType]map[string]bool{
		Any: {
			"*.*": true,
		},
		Video: {
			".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true,
			".flv": true, ".webm": true, ".m4v": true, ".mpg": true, ".mpeg": true,
			".ts": true,
		},
		Image: {
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
			".tiff": true, ".webp": true, ".svg": true, ".raw": true, ".heic": true,
			".ico": true,
		},
		Archive: {
			".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true,
			".bz2": true, ".xz": true, ".iso": true, ".tgz": true, ".tbz2": true,
		},
		Documents: {
			".docx": true, ".doc": true, ".pdf": true, ".txt": true, ".rtf": true,
			".odt": true, ".xlsx": true, ".xls": true, ".pptx": true, ".ppt": true,
			".csv": true, ".md": true, ".pages": true,
		},
	}
)

//NewFileFinder initializes a new FileFinder object
func NewFileFinder() *FileFinder {
	return &FileFinder{
		Results: make(map[string][]string),
	}
}
