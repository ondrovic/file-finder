package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"file-finder/internal/types"
	"file-finder/internal/utils"
)

var rootCmd = &cobra.Command{
	Use:   "file-finder [directory]",
	Short: "Find files of specified size and type",
	Args:  cobra.ExactArgs(1),
	Run:   run,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().BoolP("delete", "d", false, "Delete found files\n")
	rootCmd.Flags().StringP("size", "s", "315 KB", "File size to search for (1 KB, 1 MB, 1 GB)\n")
	rootCmd.Flags().IntP("type", "t", int(types.Video), "File type to search for (1: Any, 2: Video, 3: Image, 4: Archive, 5: Documents)\n")
	rootCmd.Flags().IntP("operator", "o", int(types.EqualToType), "Operator to apply on file size (1: Equal, 2: Greater Than, 3: Greater Than Or Equal To, 4: Less Than, 5: Less Than Or Equal To)\n")

	viper.BindPFlag("delete", rootCmd.Flags().Lookup("delete"))
	viper.BindPFlag("size", rootCmd.Flags().Lookup("size"))
	viper.BindPFlag("type", rootCmd.Flags().Lookup("type"))
	viper.BindPFlag("operator", rootCmd.Flags().Lookup("operator"))

	viper.SetDefault("operator", int(types.EqualToType)) // Set the default value for the "operator" flag

}

func initConfig() {
	viper.SetEnvPrefix("FF")
	viper.AutomaticEnv()
}

func run(cmd *cobra.Command, args []string) {
	finder := types.NewVideoFinder()
	finder.RootDir = args[0]
	finder.DeleteFlag = viper.GetBool("delete")
	finder.FileSize = viper.GetString("size")
	finder.FileType = types.FileType(viper.GetInt("type"))
	finder.OperatorType = types.OperatorType(viper.GetInt("operator"))
	Run(finder)
}

func main() {
	utils.ClearConsole()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Run(vf *types.VideoFinder) {
	fileSizeBytes, err := utils.ConvertSizeToBytes(vf.FileSize)
	if err != nil {
		pterm.Error.Printf("Error converting file size: %v\n", err)
		return
	}

	// Format the file size for logging
	fileSizeStr := utils.FormatSize(fileSizeBytes)

	fileTypeToString := utils.GetFileTypeToString(vf.FileType)
	operatorToString := utils.GetOperatorToString(vf.OperatorType)

	pterm.Info.Printf("Searching for files of type %v and size %s %s...\n", fileTypeToString, operatorToString, fileSizeStr)

	err = FindFiles(vf)
	if err != nil {
		pterm.Error.Printf("Error walking the path %v: %v\n", vf.RootDir, err)
		return
	}

	DisplayResults(vf)

	if vf.DeleteFlag {
		DeleteFiles(vf)
	}
}

func FindFiles(vf *types.VideoFinder) error {
	// Convert file size string to bytes
	fileSize, err := utils.ConvertSizeToBytes(vf.FileSize)
	if err != nil {
		return err
	}

	// Estimate the total number of files
	count, err := utils.FileCount(vf.RootDir, vf.FileType)
	if err != nil {
		return err
	}
	// Create a progress bar with the estimated total
	progressbar, _ := pterm.DefaultProgressbar.WithTotal(count).Start()

	// Walk through the files in the directory
	err = filepath.Walk(vf.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && utils.IsFileOfType(filepath.Ext(path), vf.FileType) {

			sizeMatch := utils.GetOperatorSizeMatches(vf.OperatorType, fileSize, info.Size())

			if sizeMatch {
				dir := filepath.Dir(path)
				vf.Results[dir] = append(vf.Results[dir], path)

				// Update the progress bar
				progressbar.Increment()
			}
		}
		return nil
	})

	progressbar.Stop()
	return err
}

func DisplayResults(vf *types.VideoFinder) {
	// Prepare results slice
	results := make([]types.DirectoryResult, 0, len(vf.Results))
	totalFiles := 0

	for dir, files := range vf.Results {
		fileCount := len(files)
		totalFiles += fileCount
		results = append(results, types.DirectoryResult{
			Directory: dir,
			Count:     fileCount,
		})
	}

	if totalFiles > 0 {
		// Render the results table
		utils.RenderResultsTable(results, totalFiles)
	} else {
		pterm.Info.Printf("%d files found matching criteria\n", totalFiles)
	}
}

func DeleteFiles(vf *types.VideoFinder) {
	result, _ := pterm.DefaultInteractiveContinue.Show("Are you sure you want to delete these files?")
	if result != "y" {
		pterm.Info.Println("Deletion cancelled.")
		return
	}

	spinner, _ := pterm.DefaultSpinner.Start("Deleting files...")
	deletedCount := 0

	for _, files := range vf.Results {
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
