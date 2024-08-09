package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"file-finder/internal/types"

	commonTypes "github.com/ondrovic/common/types"
	commonUtils "github.com/ondrovic/common/utils"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pterm/pterm"
)

var semaphore = make(chan struct{}, runtime.NumCPU())

// #region Public Functions

// FindAndDisplayFiles gathers the results and displays them
func FindAndDisplayFiles(ff types.FileFinder, targetSize int64, toleranceSize float64, detailedListing bool) error {
	results, count, err := getFiles(ff.RootDir, ff.FileType, targetSize, ff.OperatorType, toleranceSize, detailedListing)
	if err != nil {
		return err
	}

	progressbar, _ := pterm.DefaultProgressbar.WithTotal(count).Start()

	if !detailedListing {
		ff.Results = results.(map[string][]string)
		for i := 0; i < count; i++ {
			progressbar.Increment()
		}
		results = processResults(ff.Results)
	} else {
		for i := 0; i < count; i++ {
			progressbar.Increment()
		}
	}

	progressbar.Stop()

	if count > 0 {
		renderResultsToTable(results, count, detailedListing)
	} else {
		pterm.Info.Printf("%d results found matching criteria\n", count)
	}

	return nil
}

// DeleteFiles
func DeleteFiles(ff types.FileFinder) {
	result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to delete these files?")
	if !result {
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
// getFiles handles getting the files based on the criteria
func getFiles(dir string, fileType commonTypes.FileType, targetSize int64, operatorType commonTypes.OperatorType, toleranceSize float64, detailed bool) (interface{}, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, 0, err
	}

	var detailedResults []types.EntryResults
	results := make(map[string][]string)
	var totalCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, entry := range entries {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			if entry.IsDir() {
				semaphore <- struct{}{}
				subResult, subCount, err := getFiles(path, fileType, targetSize, operatorType, toleranceSize, detailed)
				<-semaphore
				if err != nil {
					return
				}
				mu.Lock()
				if detailed {
					detailedResults = append(detailedResults, subResult.([]types.EntryResults)...)
				} else {
					for dir, files := range subResult.(map[string][]string) {
						results[dir] = append(results[dir], files...)
					}
				}
				totalCount += subCount
				mu.Unlock()
			} else {
				if commonUtils.IsExtensionValid(fileType, path) {
					info, err := entry.Info()
					if err != nil {
						return
					}

					size := info.Size()

					matches, err := commonUtils.GetOperatorSizeMatches(operatorType, targetSize, toleranceSize, size)

					if err != nil {
						pterm.Error.Printf("Error calculating tolerances: %v\n", err)
					}

					if matches {
						mu.Lock()
						if detailed {
							detailedResults = append(detailedResults, types.EntryResults{
								Directory: dir,
								FileName:  entry.Name(),
								FileSize:  commonUtils.FormatSize(size),
							})
						} else {
							results[dir] = append(results[dir], path)
						}
						totalCount++
						mu.Unlock()
					}
				}
			}
		}(filepath.Join(dir, entry.Name()))
	}

	wg.Wait()

	if detailed {
		return detailedResults, totalCount, nil
	}
	return results, totalCount, nil
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

// renderResultsToTable renders the results into a formatted table
func renderResultsToTable(results interface{}, totalCount int, detailedListing bool) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	if detailedListing {
		t.AppendHeader(table.Row{"Directory", "FileName", "FileSize"})
		for _, result := range results.([]types.EntryResults) {
			t.AppendRow(table.Row{
				commonUtils.FormatPath(result.Directory, runtime.GOOS),
				result.FileName,
				result.FileSize,
			})
		}
	} else {
		t.AppendHeader(table.Row{"Directory", "Count"})
		for _, result := range results.([]types.DirectoryResult) {
			t.AppendRow(table.Row{
				commonUtils.FormatPath(result.Directory, runtime.GOOS),
				result.Count,
			})
		}
	}

	t.AppendFooter(table.Row{"Total", totalCount})
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}

// #endregion
