package utils

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"unicode"

	"file-finder/internal/types"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pterm/pterm"
)

// #region Public Functions

// ClearConsole clears terminal based on operating system
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

// ToFileType converts a string to FileType, case-insensitive
func ToFileType(s string) types.FileType {
	switch strings.ToLower(s) {
	case "any":
		return types.Any
	case "video":
		return types.Video
	case "image":
		return types.Image
	case "archive":
		return types.Archive
	case "documents":
		return types.Documents
	default:
		return ""
	}
}

// ToOperatorType converts a string to OperatorType, case-insensitive
func ToOperatorType(s string) types.OperatorType {
	switch strings.ToLower(s) {
	case "equal to", "equalto", "equal", "==":
		return types.EqualTo
	case "greater than", "greaterthan", ">":
		return types.GreaterThan
	case "greater than or equal to", "greaterthanorequalto", ">=":
		return types.GreaterThanEqualTo
	case "less than", "lessthan", "<":
		return types.LessThan
	case "less than or equal to", "lessthanorequalto", "<=":
		return types.LessThanEqualTo
	default:
		return ""
	}
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
	if unitStr == "" || numStr == "" {
		return 0, errors.New("invalid size format")
	}

	// Normalize the unit string to uppercase
	unitStr = strings.ToUpper(unitStr)

	// Parse the numeric part
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}

	// Find the matching unit and convert to bytes
	for _, unit := range types.Units {
		if unit.Label == unitStr {
			return int64(num * float64(unit.Size)), nil
		}
	}

	return 0, errors.New("invalid size unit")
}

// FormatSize formats size to human readable
func FormatSize(bytes int64) string {
	for _, unit := range types.Units {
		if bytes >= unit.Size {
			value := float64(bytes) / float64(unit.Size)
			// Round the value to two decimal places
			roundedValue := math.Round(value*100) / 100
			return fmt.Sprintf("%.2f %s", roundedValue, unit.Label)
		}
	}

	return "0 B"
}

// CalculateTolerancePercentage calculates the tolerance percentage
func CalculateTolerancePercentage(tolerance float64, precision int) string {
	percentValue := tolerance * 100
	format := fmt.Sprintf("%%.%df%%%%", precision)
	return fmt.Sprintf(format, percentValue)
}

// CalculateToleranceBytes calculates the tolerance size in bytes
func CalculateToleranceBytes(sizeStr string, tolerance float64) (int64, error) {
	fileSize, err := ConvertSizeToBytes(sizeStr)
	if err != nil {
		return 0, err
	}

	toleranceFactor := tolerance / 100.0
	newSize := float64(fileSize) * (1 + toleranceFactor)

	nSize := int64(newSize)

	return nSize, nil
}

// FindAndDisplay gathers the results and displays them
func FindAndDisplay(ff types.FileFinder, targetSize int64, toleranceSize int64, detailed bool) error {
    var results interface{}
    var err error
    var count int

    if detailed {
        results, err = entryCount(ff.RootDir, ff.FileType, targetSize, ff.OperatorType, toleranceSize)
        count = len(results.([]types.EntryResults))
    } else {
        count, err = fileCount(ff.RootDir, ff.FileType, targetSize, ff.OperatorType, toleranceSize)
    }

    if err != nil {
        return err
    }

    progressbar, _ := pterm.DefaultProgressbar.WithTotal(count).Start()

    if !detailed {
        err = filepath.Walk(ff.RootDir, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }
            if !info.IsDir() && isFileOfType(filepath.Ext(path), ff.FileType) {
                size := info.Size()
                if getOperatorSizeMatches(ff.OperatorType, targetSize, toleranceSize, size) {
                    dir := filepath.Dir(path)
                    ff.Results[dir] = append(ff.Results[dir], path)
                    progressbar.Increment()
                }
            }
            return nil
        })
        results = processResults(ff.Results)
    } else {
        i := 0
        for i < count {
            progressbar.Increment()
            i++
        }
    }

    progressbar.Stop()

    if count > 0 {
        renderResults(results, count, detailed)
    } else {
        pterm.Info.Printf("%d results found matching criteria\n", count)
    }

    return err
}

// DeleteFiles
func DeleteFiles(ff types.FileFinder) {
	result, _ := pterm.DefaultInteractiveContinue.Show("Are you sure you want to delete these files?")
	if result != "y" {
		pterm.Info.Println("Deletion cancelled.")
		return
	}

	spinner, _ := pterm.DefaultSpinner.Start("Deleting files...")
	deletedCount := 0

	for _, files := range ff.Results {
		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				pterm.Error.Printf("Error deleting %s: %v\n", file, err)
			} else {
				deletedCount++
			}
		}
	}

	spinner.Success(fmt.Sprintf("Deleted %d files.", deletedCount))
}

// #endregion

// #region Internal Functions

// isExtensionValid checks if the file's extension is allowed for a given file type.
func isExtensionValid(fileType types.FileType, path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	extensions, exists := types.FileExtensions[fileType]
	if !exists {
		return false
	}

	// Check for wildcard entry (Any)
	if _, found := extensions["*.*"]; found {
		return true
	}

	// Check if the file extension is explicitly allowed
	return extensions[ext]
}

// entryCount
func entryCount(root string, fileType types.FileType, targetSize int64, operatorType types.OperatorType, toleranceSize int64) ([]types.EntryResults, error) {
	count, err := countEntries(root, fileType, targetSize, operatorType, toleranceSize)
	return count, err
}

// countEntries
func countEntries(dir string, fileType types.FileType, targetSize int64, operatorType types.OperatorType, toleranceSize int64) ([]types.EntryResults, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var results []types.EntryResults

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			subResults, err := countEntries(path, fileType, targetSize, operatorType, toleranceSize)
			if err != nil {
				return nil, err
			}
			results = append(results, subResults...)
		} else {
			if isExtensionValid(fileType, path) {
				info, err := entry.Info()
				if err != nil {
					return nil, err
				}

				size := info.Size()

				if getOperatorSizeMatches(operatorType, targetSize, toleranceSize, size) {
					results = append(results, types.EntryResults{
						Directory: dir,
						FileName:  entry.Name(),
						FileSize:  FormatSize(size),
					})
				}
			}
		}
	}
	return results, nil
}

// fileCount
func fileCount(root string, fileType types.FileType, targetSize int64, operatorType types.OperatorType, toleranceSize int64) (int, error) {
	count, err := countFiles(root, fileType, targetSize, operatorType, toleranceSize)
	return count, err
}

// countFiles
func countFiles(dir string, fileType types.FileType, targetSize int64, operatorType types.OperatorType, toleranceSize int64) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}

	var count int

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			// Recursively count files in subdirectories and accumulate the count
			subdirCount, err := countFiles(path, fileType, targetSize, operatorType, toleranceSize)
			if err != nil {
				return 0, err
			}
			count += subdirCount
		} else {
			if isExtensionValid(fileType, path) {
				// Get file info to check size
				info, err := entry.Info()
				if err != nil {
					return 0, err
				}

				size := info.Size()

				if getOperatorSizeMatches(operatorType, targetSize, toleranceSize, size) {
					count++
				}
			}
		}
	}
	return count, nil
}

// getOperatorSizeMatches determines whether or not a file matches the size or tolerance size
func getOperatorSizeMatches(operator types.OperatorType, fileSize int64, toleranceSize int64, infoSize int64) bool {
	switch operator {
	case types.EqualTo:
		return (infoSize == fileSize || infoSize == toleranceSize)
	case types.LessThan:
		return (infoSize < fileSize || infoSize < toleranceSize)
	case types.LessThanEqualTo:
		return (infoSize <= fileSize || infoSize <= toleranceSize)
	case types.GreaterThan:
		return (infoSize > fileSize || infoSize > toleranceSize)
	case types.GreaterThanEqualTo:
		return (infoSize >= fileSize || infoSize >= toleranceSize)
	default:
		return (infoSize == fileSize || infoSize == toleranceSize)
	}
}

// isFileOfType checks if a file extension matches the given file type
func isFileOfType(ext string, fileType types.FileType) bool {
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

// formatPath formats the path output based on operating system
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

// processResults processes the results
func processResults(results map[string][]string) []types.DirectoryResult {
	var processedResults []types.DirectoryResult
	for dir, files := range results {
		processedResults = append(processedResults, types.DirectoryResult{
			Directory: dir,
			Count:     len(files),
		})
	}
	return processedResults
}

// renderResults
func renderResults(results interface{}, totalCount int, detailed bool) {
	if detailed {
		renderResultsTableEntry(results.([]types.EntryResults), totalCount)
	} else {
		renderResultsTable(results.([]types.DirectoryResult), totalCount)
	}
}

// renderResultsTableEntry renders the results for the detailed entries results
func renderResultsTableEntry(results []types.EntryResults, totalCount int) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Directory", "FileName", "FileSize"})
	for _, result := range results {
		t.AppendRow(table.Row{
			formatPath(result.Directory),
			result.FileName,
			result.FileSize,
		})
	}
	t.AppendFooter(table.Row{"Total", totalCount})
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}

// renderResultsTable renders the non detailed results
func renderResultsTable(results []types.DirectoryResult, totalCount int) {
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

// #endregion
