package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"

	"file-finder/internal/types"

	commonUtils "github.com/ondrovic/common/utils"
	commonFormatters "github.com/ondrovic/common/utils/formatters"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/pterm/pterm"
)

var (
	semaphore = make(chan struct{})
)

// FindAndDisplayFiles gathers the results and displays them
func FindAndDisplayFiles(ff types.FileFinder) (interface{}, error) {
	results, count, size, err := getFiles(ff)
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
		renderResultsToTable(results, count, size, ff)
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
func getFiles(ff types.FileFinder) (interface{}, int, int64, error) {
	entries, err := os.ReadDir(ff.RootDirectory)
	if err != nil {
		return nil, 0, 0, err
	}

	// Initialize variables
	var mu sync.Mutex
	var wg sync.WaitGroup
	results := make(map[string][]string)
	var detailedResults []types.EntryResult
	var totalCount int
	var totalFileSize int64

	// Handle file size filter conversion
	fileSize, err := convertFileSizeFilter(ff.FileSizeFilter)
	if err != nil {
		return nil, 0, 0, err
	}

	// Process each directory entry
	semaphore = make(chan struct{}, runtime.NumCPU())
	for _, entry := range entries {
		wg.Add(1)
		go func(entry os.DirEntry) {
			defer wg.Done()
			path := filepath.Join(ff.RootDirectory, entry.Name())
			if entry.IsDir() {
				processDirectory(path, ff, &detailedResults, &results, &totalCount, &totalFileSize, &mu, &wg, semaphore)
			} else {
				processFile(entry, path, ff, fileSize, &detailedResults, &results, &totalCount, &totalFileSize, &mu)
			}
		}(entry)
	}

	wg.Wait()

	if ff.DisplayDetailedResults {
		return detailedResults, totalCount, totalFileSize, nil
	}
	return results, totalCount, 0, nil
}

// convertFileSizeFilter converts the file size filter string to bytes
func convertFileSizeFilter(fileSizeFilter string) (int64, error) {
	if fileSizeFilter == "" {
		return 0, nil
	}
	return commonUtils.ConvertStringSizeToBytes(fileSizeFilter)
}

func processDirectory(path string, ff types.FileFinder, detailedResults *[]types.EntryResult, results *map[string][]string, totalCount *int, totalFileSize *int64, mu *sync.Mutex, wg *sync.WaitGroup, semaphore chan struct{}) {
	semaphore <- struct{}{}        // Acquire semaphore
	defer func() { <-semaphore }() // Release semaphore

	subFF := ff
	subFF.RootDirectory = path
	subResult, subCount, subSize, err := getFiles(subFF)
	if err != nil {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if ff.DisplayDetailedResults {
		*detailedResults = append(*detailedResults, subResult.([]types.EntryResult)...)
		*totalFileSize += subSize
	} else {
		for dir, files := range subResult.(map[string][]string) {
			(*results)[dir] = append((*results)[dir], files...)
		}
	}
	*totalCount += subCount
}

// processFile handles processing of a single file
func processFile(entry os.DirEntry, path string, ff types.FileFinder, fileSize int64, detailedResults *[]types.EntryResult, results *map[string][]string, totalCount *int, totalFileSize *int64, mu *sync.Mutex) {
	if !commonUtils.IsExtensionValid(ff.FileTypeFilter, path) {
		return
	}

	info, err := entry.Info()
	if err != nil {
		return
	}

	size := info.Size()

	// Apply file size filter if necessary
	if ff.FileSizeFilter != "" && !applyFileSizeFilter(ff, size, fileSize) {
		return
	}

	// Apply file name filter if necessary
	if !applyFileNameFilter(ff, entry.Name()) {
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if ff.DisplayDetailedResults {
		*detailedResults = append(*detailedResults, types.EntryResult{
			Directory: ff.RootDirectory,
			FileName:  entry.Name(),
			FileSize:  commonFormatters.FormatSize(size),
		})
		*totalFileSize += size
	} else {
		(*results)[ff.RootDirectory] = append((*results)[ff.RootDirectory], path)
	}
	*totalCount++
}

// applyFileSizeFilter checks if a file matches the size criteria
func applyFileSizeFilter(ff types.FileFinder, size, fileSize int64) bool {
	sizeMatches, err := commonUtils.GetOperatorSizeMatches(ff.OperatorTypeFilter, fileSize, ff.ToleranceSize, size)
	if err != nil {
		pterm.Error.Println(err)
		return false
	}
	return sizeMatches
}

// applyFileNameFilter checks if a file matches the name criteria
func applyFileNameFilter(ff types.FileFinder, fileName string) bool {
	if ff.FileNameFilter == "" {
		return true
	}

	lowerEntryName, err := commonFormatters.ToLower(fileName)
	if err != nil {
		pterm.Error.Println(err)
		return false
	}
	lowerFileNameFilter, err := commonFormatters.ToLower(ff.FileNameFilter)
	if err != nil {
		pterm.Error.Println(err)
		return false
	}

	return strings.Contains(lowerEntryName, lowerFileNameFilter)
}

func deleteEntryResults(entries []types.EntryResult) (int, []string) {
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
	case []types.EntryResult:
		deletedFileCount, directoriesToRemove = deleteEntryResults(v)
	// Part of the bug related to the func above
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
		// for debugging since you cannot interact
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

func formatResultHyperLink(link, txt string) string {
	text.EnableColors()

	link = commonFormatters.FormatPath(link, runtime.GOOS)
	txt = text.FgGreen.Sprint(txt)

	return text.Hyperlink(link, txt)
}

func renderResultsToTable(results interface{}, totalCount int, totalFileSize int64, ff types.FileFinder) {
	t := table.Table{}

	// Determine header and footer based on the type of results
	var header table.Row
	var footer table.Row
	switch results.(type) {
	case []types.DirectoryResult:
		header = table.Row{"Directory", "Count"}
		footer = table.Row{"Total", pterm.Sprintf("%v", totalCount)}
	case []types.EntryResult:
		header = table.Row{"Directory", "FileName", "FileSize"}
		footer = table.Row{"Total", pterm.Sprintf("%v", totalCount), pterm.Sprintf("%v", commonFormatters.FormatSize(totalFileSize))}
	default:
		return // Exit if results type is not supported
	}

	t.AppendHeader(header)

	// Append rows based on the display mode
	switch results := results.(type) {
	case []types.DirectoryResult:
		for _, result := range results {
			t.AppendRow(table.Row{
				formatResultHyperLink(result.Directory, result.Directory),
				pterm.Sprintf("%v", result.Count),
			})
		}
	case []types.EntryResult:
		if ff.DisplayDetailedResults {
			for _, result := range results {
				newLink := pterm.Sprintf("%s/%s", result.Directory, result.FileName)
				t.AppendRow(table.Row{
					formatResultHyperLink(result.Directory, result.Directory),
					formatResultHyperLink(newLink, result.FileName),
					result.FileSize,
				})
			}
		}
	}

	t.AppendFooter(footer)

	t.SetStyle(table.StyleColoredDark)
	t.SetOutputMirror(os.Stdout)
	t.Render()
}
