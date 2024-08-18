package utils

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	// "fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"

	"file-finder/internal/types"

	// commonTypes "github.com/ondrovic/common/types"
	commonUtils "github.com/ondrovic/common/utils"
	commonFormatters "github.com/ondrovic/common/utils/formatters"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/pterm/pterm"
)

var semaphore = make(chan struct{}, runtime.NumCPU())

// FindAndDisplayFiles gathers the results and displays them
func FindAndDisplayFiles(ff types.FileFinder) (interface{}, error) {
	results, count, err := getFiles(ff)
	if err != nil {
		return nil, err
	}

	progressbar, _ := pterm.DefaultProgressbar.WithTotal(count).WithRemoveWhenDone(true).Start()

	if !ff.DisplayDetailedResults {
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
		renderResultsToTable(results, count, ff)
	} else {
		pterm.Info.Printf("%d results found matching criteria\n", count)
	}

	return results, nil
}

// getResultsCount returns the count of elements in a slice or array
func getResultsCount(results interface{}) (int, error) {
	val := reflect.ValueOf(results)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return 0, errors.New("results is not a slice or array")
	}
	return val.Len(), nil
}

// getFiles handles getting the files based on the criteria
func getFiles(ff types.FileFinder) (interface{}, int, error) {

	entries, err := os.ReadDir(ff.RootDirectory)
	if err != nil {
		return nil, 0, err
	}

	var detailedResults []types.EntryResults
	results := make(map[string][]string)
	var totalCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// convert filter to bytes
	var fileSize int64
	if ff.FileSizeFilter != "" {
		var err error
		fileSize, err = commonUtils.ConvertStringSizeToBytes(ff.FileSizeFilter)
		if err != nil {
			return nil, 0, err
		}
	}

	for _, entry := range entries {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			if entry.IsDir() {
				semaphore <- struct{}{}
				subFF := ff
				subFF.RootDirectory = path
				subResult, subCount, err := getFiles(subFF)
				<-semaphore
				if err != nil {
					return
				}
				mu.Lock()
				if ff.DisplayDetailedResults {
					detailedResults = append(detailedResults, subResult.([]types.EntryResults)...)
				} else {
					for dir, files := range subResult.(map[string][]string) {
						results[dir] = append(results[dir], files...)
					}
				}
				totalCount += subCount
				mu.Unlock()
			} else {
				if commonUtils.IsExtensionValid(ff.FileTypeFilter, path) {
					info, err := entry.Info()
					if err != nil {
						return
					}

					size := info.Size()

					// Apply FileSize filter only if it's present
					var sizeMatches bool
					if ff.FileSizeFilter != "" {
						sizeMatches, err = commonUtils.GetOperatorSizeMatches(ff.OperatorTypeFilter, fileSize, ff.ToleranceSize, size)
						if err != nil {
							// pterm.Error.Printf("Error getting size matches: %v\n", err)
							pterm.Error.Println(err)
							return
						}
						if !sizeMatches {
							return
						}
					}

					// convert ToLower
					lowerEntryName, err := commonFormatters.ToLower(entry.Name())
					if err != nil {
						pterm.Error.Println(err)
						return
					}
					lowerFileNameFilter, err := commonFormatters.ToLower(ff.FileNameFilter)
					if err != nil {
						pterm.Error.Println(err)
						return
					}
					// Apply FileNameFilter if present
					if ff.FileNameFilter != "" && !strings.Contains(lowerEntryName, lowerFileNameFilter) {
						return
					}

					mu.Lock()
					if ff.DisplayDetailedResults {
						detailedResults = append(detailedResults, types.EntryResults{
							Directory: ff.RootDirectory,
							FileName:  entry.Name(),
							FileSize:  commonFormatters.FormatSize(size),
						})
					} else {
						results[ff.RootDirectory] = append(results[ff.RootDirectory], path)
					}
					totalCount++
					mu.Unlock()
				}
			}
		}(filepath.Join(ff.RootDirectory, entry.Name()))
	}

	wg.Wait()

	if ff.DisplayDetailedResults {
		return detailedResults, totalCount, nil
	}
	return results, totalCount, nil
}

func deleteEntryResults(entries []types.EntryResults) (int, []string) {
	var deletedCount int
	var directoriesToRemove []string
	for _, entry := range entries {
		directoriesToRemove = append(directoriesToRemove, entry.Directory)
		filePath := filepath.Join(entry.Directory, entry.FileName)
		if err := os.Remove(filePath); err != nil {
			pterm.Error.Printf("Error deleting %s: %v\n", filePath, err)
		} else {
			deletedCount++
		}
	}
	return deletedCount, directoriesToRemove
}


// BUG: when doing the directory result it doesn't list the files so you end up deleting the entire directory of files ;-(
//      going to think on how I want to do this, but for now I am just doing to disable -r unless -d is used
// func deleteDirectoryResults(dirResults []types.DirectoryResult) (int, []string) {
// 	var deletedCount int
// 	var directoriesToRemove []string

// 	for _, dirResult := range dirResults {
// 		// Add the base directory to the list of directories to remove
// 		directoriesToRemove = append(directoriesToRemove, dirResult.Directory)

// 		// bug: I think we need to make sure the directory is empty before deleting it
// 		err := filepath.Walk(dirResult.Directory, func(path string, info os.FileInfo, err error) error {
// 			if err != nil {
// 				return err
// 			}
// 			if !info.IsDir() {
// 				if err := os.Remove(path); err != nil {
// 					pterm.Error.Printf("Error deleting file %s: %v\n", path, err)
// 				} else {
// 					deletedCount++
// 				}
// 			} else {
// 				directoriesToRemove = append(directoriesToRemove, path)
// 			}
// 			return nil
// 		})

// 		if err != nil {
// 			pterm.Error.Printf("Error walking directory %s: %v\n", dirResult.Directory, err)
// 		}
// 	}

// 	return deletedCount, directoriesToRemove
// }


func deleteFileBasedOnResults(results interface{}) error {
	spinner, _ := pterm.DefaultSpinner.Start("Deleting files and directories...")
	defer spinner.Stop()

	var deletedFileCount int
	var directoriesToRemove []string

	switch v := results.(type) {
	case []types.EntryResults:
		deletedFileCount, directoriesToRemove = deleteEntryResults(v)
	// case []types.DirectoryResult:
	// 		deletedFileCount, directoriesToRemove = deleteDirectoryResults(v)
	default:
		return fmt.Errorf("invalid data format: expected []EntryResults or []DirectoryResult, got %T", results)
	}

	// Sort and filter directories
	directoriesToRemove = sortAndFilterDirs(directoriesToRemove)

	// Delete empty directories
	deletedDirCount, err := deleteEmptyDirectories(directoriesToRemove)
	if err != nil {
		return fmt.Errorf("error deleting empty directories: %w", err)
	}

	spinner.Success(fmt.Sprintf("Deleted %d files and %d directories.", deletedFileCount, deletedDirCount))
	return nil
}

func deleteEmptyDirectories(directories []string) (int, error) {
	removedCount := 0
	for _, dir := range directories {
		empty, err := isDirEmpty(dir)
		if err != nil {
			if !os.IsNotExist(err) {
				pterm.Error.Printf("Error checking if directory is empty %s: %v\n", dir, err)
			}
			continue
		}
		if empty {
			if err := os.Remove(dir); err != nil {
				if !os.IsNotExist(err) {
					pterm.Error.Printf("Error deleting directory %s: %v\n", dir, err)
				}
			} else {
				removedCount++
			}
		}
	}
	return removedCount, nil
}

func sortAndFilterDirs(directories []string) []string {
	dirSet := make(map[string]struct{})
	for _, dir := range directories {
		dirSet[dir] = struct{}{}
		// Add all parent directories
		for d := dir; d != filepath.Dir(d); d = filepath.Dir(d) {
			dirSet[d] = struct{}{}
		}
	}

	uniqueDirs := make([]string, 0, len(dirSet))
	for dir := range dirSet {
		uniqueDirs = append(uniqueDirs, dir)
	}

	// Sort directories by depth (deepest first) and then by length
	sort.SliceStable(uniqueDirs, func(i, j int) bool {
		depthI := strings.Count(uniqueDirs[i], string(os.PathSeparator))
		depthJ := strings.Count(uniqueDirs[j], string(os.PathSeparator))
		if depthI != depthJ {
			return depthI > depthJ
		}
		return len(uniqueDirs[i]) > len(uniqueDirs[j])
	})

	return uniqueDirs
}

func isDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// DeleteFiles
func DeleteFiles(results interface{}) {

	resultCount, err := getResultsCount(results)
	if err != nil {
		pterm.Error.Printf("Error getting results count %s: %v\n", results, err)
	}

	if resultCount > 0 {
		// Confirm deletion with the user (you may want to uncomment this if needed)
		result, _ := pterm.DefaultInteractiveConfirm.Show("Are you sure you want to delete these files?")
		// result := true
		if !result {
			pterm.Info.Println("Deletion cancelled.")
			return
		}

		deleteFileBasedOnResults(results)
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

// renderResultsToTable renders the results into a formatted table
func renderResultsToTable(results interface{}, totalCount int, ff types.FileFinder) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	if ff.DisplayDetailedResults {
		t.AppendHeader(table.Row{"Directory", "FileName", "FileSize"})
		for _, result := range results.([]types.EntryResults) {
			t.AppendRow(table.Row{
				commonFormatters.FormatPath(result.Directory, runtime.GOOS),
				result.FileName,
				result.FileSize,
			})
		}
	} else {
		t.AppendHeader(table.Row{"Directory", "Count"})
		for _, result := range results.([]types.DirectoryResult) {
			t.AppendRow(table.Row{
				commonFormatters.FormatPath(result.Directory, runtime.GOOS),
				result.Count,
			})
		}
	}

	t.AppendFooter(table.Row{"Total", totalCount})
	t.SetStyle(table.StyleColoredDark)
	t.Render()
}
