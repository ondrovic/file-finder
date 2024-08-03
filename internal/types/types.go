package types

// FileType represents different types of files
type FileType int

// OperatorTypes represents different file size operations
type OperatorType int


const (
	_ = iota			// 0
	Any FileType = iota // 1
	Video               // 2
	Image               // 3
	Archive             // 4
	Documents           // 5
)


const (
	_ = iota						// 0
	EqualToType OperatorType = iota // 1
	GreaterThanType                 // 2
	GreaterThanEqualToType          // 3
	LessThanType                    // 4
	LessThanEqualToType             // 5
)

// VideoFinder struct fields should be exported (capitalized) to be accessible from the main package
type VideoFinder struct {
	RootDir      string
	DeleteFlag   bool
	FileSize     string
	FileType     FileType
	OperatorType OperatorType
	Results      map[string][]string
}

type DirectoryResult struct {
	Directory string
	Count     int
}

var (
	FileExtensions = map[FileType]map[string]bool{
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

func NewVideoFinder() *VideoFinder {
	return &VideoFinder{
		FileType:     Video,
		OperatorType: OperatorType(EqualToType),
		Results:      make(map[string][]string),
	}
}
