package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"file-finder/internal/types"

	"github.com/jedib0t/go-pretty/v6/table"
)

func ClearConsole() {
	var clearCmd *exec.Cmd

	switch runtime.GOOS {
	case "linux", "darwin":
		clearCmd = exec.Command("clear")
	case "windows":
		clearCmd = exec.Command("cmd", "/c", "cls")
	default:
		fmt.Println("Unsupported platform")
		return
	}

	clearCmd.Stdout = os.Stdout
	clearCmd.Run()
}

// ConvertSizeToBytes converts a size string with a unit to bytes.
func ConvertSizeToBytes(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, errors.New("size cannot be empty")
	}

	// Separate the numeric part and the unit part
	var numStr, unitStr string
	for i, r := range sizeStr {
		if unicode.IsLetter(r) {
			numStr = strings.TrimSpace(sizeStr[:i])
			unitStr = strings.TrimSpace(sizeStr[i:])
			break
		}
	}

	// If no unit was found, return an error
	if unitStr == "" {
		return 0, errors.New("invalid size format")
	}

	// Normalize the unit string to lowercase
	unitStr = strings.ToLower(unitStr)

	// Define the units and their corresponding byte multipliers
	units := map[string]int64{
		"b":  1,
		"kb": 1024,
		"mb": 1024 * 1024,
		"gb": 1024 * 1024 * 1024,
		"pb": 1024 * 1024 * 1024 * 1024,
	}

	// Parse the numeric part
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	// Get the multiplier for the unit
	multiplier, ok := units[unitStr]
	if !ok {
		return 0, errors.New("invalid size unit")
	}

	// Convert to bytes
	return int64(num * float64(multiplier)), nil
}

// FormatSize converts bytes to a human-readable size string.
func FormatSize(bytes int64) string {
	units := []struct {
		label string
		size  int64
	}{
		{"PB", 1024 * 1024 * 1024 * 1024},
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KB", 1024},
		{"B", 1},
	}

	for _, unit := range units {
		if bytes >= unit.size {
			value := float64(bytes) / float64(unit.size)
			return fmt.Sprintf("%.2f %s", value, unit.label)
		}
	}

	return "0 B"
}

func FileCount(root string, fileType types.FileType) (int, error) {
	var count int
	err := countFiles(root, fileType, &count)
	return count, err
}

func countFiles(dir string, fileType types.FileType, count *int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	extensions := types.FileExtensions[fileType]

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := countFiles(path, fileType, count); err != nil {
				return err
			}
		} else {
			ext := strings.ToLower(filepath.Ext(path))
			if extensions[ext] {
				*count++
			}
		}
	}
	return nil
}

func GetOperatorSizeMatches(operator types.OperatorType, fileSize int64, infoSize int64) bool {
	switch operator {
	case types.EqualToType:
		return infoSize == fileSize
	case types.LessThanType:
		return infoSize < fileSize
	case types.LessThanEqualToType:
		return infoSize <= fileSize
	case types.GreaterThanType:
		return infoSize > fileSize
	case types.GreaterThanEqualToType:
		return infoSize >= fileSize
	default:
		return infoSize == fileSize
	}
}

func GetOperatorToString(operator types.OperatorType) string {
	switch operator {
	case types.EqualToType:
		return "equal to"
	case types.LessThanType:
		return "less than"
	case types.LessThanEqualToType:
		return "less than or equal to"
	case types.GreaterThanType:
		return "greater than"
	case types.GreaterThanEqualToType:
		return "greater than or equal to"
	default:
		return "equal to"
	}
}

func GetFileTypeToString(fileType types.FileType) string {
	fileTypeToString := map[types.FileType]string{
		types.Any:       "Any",
		types.Video:     "Video",
		types.Image:     "Image",
		types.Archive:   "Archive",
		types.Documents: "Documents",
	}

	return fileTypeToString[fileType]
}

// IsFileOfType checks if a file extension matches the given file type
func IsFileOfType(ext string, fileType types.FileType) bool {
	if fileType == types.Any {
		for _, extensions := range types.FileExtensions {
			if extensions[ext] {
				return true
			}
		}
		return false
	}
	return types.FileExtensions[fileType][ext]
}

func formatPath(path string) string {
	switch runtime.GOOS {
	case "windows":
		// Convert to Windows style paths (with backslashes)
		return filepath.ToSlash(path)
	case "linux", "darwin":
		// Convert to Unix style paths (with forward slashes)
		return filepath.FromSlash(path)
	default:
		// Default to Unix style paths
		return path
	}
}

func RenderResultsTable(results []types.DirectoryResult, totalCount int) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Directory", "Count"})
	for _, result := range results {
		t.AppendRow(table.Row{
			formatPath(result.Directory),
			result.Count,
		})
	}
	t.AppendFooter(table.Row{"Total", totalCount})
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}
