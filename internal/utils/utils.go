package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"file-finder/internal/types"

	commonTypes "github.com/ondrovic/common/types"
	commonUtils "github.com/ondrovic/common/utils"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pterm/pterm"
)

// #region Public Functions

// CalculateTolerancePercentage calculates the tolerance percentage
func CalculateTolerancePercentage(tolerance float64, precision int) string {
	percentValue := tolerance * 100
	format := fmt.Sprintf("%%.%df%%%%", precision)
	return fmt.Sprintf(format, percentValue)
}

// FindAndDisplay gathers the results and displays them
func FindAndDisplay(ff types.FileFinder, targetSize int64, toleranceSize int64, detailed bool) error {
	var results interface{}
	var err error
	var count int

	if detailed {
		results, err = countEntries(ff.RootDir, ff.FileType, targetSize, ff.OperatorType, toleranceSize)
		count = len(results.([]types.EntryResults))
	} else {
		count, err = countFiles(ff.RootDir, ff.FileType, targetSize, ff.OperatorType, toleranceSize)
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
			// if !info.IsDir() && isFileOfType(filepath.Ext(path), ff.FileType) {
			if !info.IsDir() && commonUtils.IsExtensionValid(ff.FileType, path) {
				size := info.Size()
				if commonUtils.GetOperatorSizeMatches(ff.OperatorType, targetSize, toleranceSize, size) {
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

// countEntries
func countEntries(dir string, fileType commonTypes.FileType, targetSize int64, operatorType commonTypes.OperatorType, toleranceSize int64) ([]types.EntryResults, error) {
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
			if commonUtils.IsExtensionValid(fileType, path) {
				info, err := entry.Info()
				if err != nil {
					return nil, err
				}

				size := info.Size()

				if commonUtils.GetOperatorSizeMatches(operatorType, targetSize, toleranceSize, size) {
					results = append(results, types.EntryResults{
						Directory: dir,
						FileName:  entry.Name(),
						FileSize:  commonUtils.FormatSize(size),
					})
				}
			}
		}
	}
	return results, nil
}

// countFiles
func countFiles(dir string, fileType commonTypes.FileType, targetSize int64, operatorType commonTypes.OperatorType, toleranceSize int64) (int, error) {
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
			if commonUtils.IsExtensionValid(fileType, path) {
				// Get file info to check size
				info, err := entry.Info()
				if err != nil {
					return 0, err
				}

				size := info.Size()

				if commonUtils.GetOperatorSizeMatches(operatorType, targetSize, toleranceSize, size) {
					count++
				}
			}
		}
	}
	return count, nil
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
			commonUtils.FormatPath(result.Directory, runtime.GOOS),
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
			commonUtils.FormatPath(result.Directory, runtime.GOOS),
			result.Count,
		})
	}
	t.AppendFooter(table.Row{"Total", totalCount})
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}

// #endregion
