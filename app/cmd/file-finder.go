package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"file-finder/internal/types"
	"file-finder/internal/utils"
)

// #region Cli Setup
var rootCmd = &cobra.Command{
	Use:   "file-finder [directory]",
	Short: "Find files of specified size and type",
	Args:  cobra.ExactArgs(1),
	Run:   run,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringP("type", "t", string(types.Video), "File type to search for (Any, Video, Image, Archive, Documents)\n")
	rootCmd.Flags().StringP("operator", "o", string(types.EqualTo), "Operator to apply on file size (Equal To, Greater Than, Greater Than Or Equal To, Less Than, Less Than Or Equal To)\n")
	rootCmd.Flags().BoolP("delete", "d", false, "Delete found files\n(default: false)")
	rootCmd.Flags().BoolP("detailed", "r", false, "Display detailed results\n(default: false)")
	rootCmd.Flags().StringP("size", "s", "315 KB", "File size to search for (1 KB, 1 MB, 1 GB)\n")
	rootCmd.Flags().Float64P("tolerance", "l", 0.01, "File size tolerance\n")

	// Bind flags with viper
	viper.BindPFlag("delete", rootCmd.Flags().Lookup("delete"))
	viper.BindPFlag("detailed", rootCmd.Flags().Lookup("detailed"))
	viper.BindPFlag("size", rootCmd.Flags().Lookup("size"))
	viper.BindPFlag("tolerance", rootCmd.Flags().Lookup("tolerance"))
	viper.BindPFlag("type", rootCmd.Flags().Lookup("type"))
	viper.BindPFlag("operator", rootCmd.Flags().Lookup("operator"))

}

func initConfig() {
	viper.SetEnvPrefix("FF")
	viper.AutomaticEnv()
}

func run(cmd *cobra.Command, args []string) {
	fileType := utils.ToFileType(viper.GetString("type"))
	operatorType := utils.ToOperatorType(viper.GetString("operator"))

	if fileType == "" {
		pterm.Error.Printf("invalid file type: %s", viper.GetString("type"))
	}
	if operatorType == "" {
		pterm.Error.Printf("invalid operator type: %s", viper.GetString("operator"))
	}

	finder := types.FileFinder{
		RootDir:          args[0],
		DeleteFlag:       viper.GetBool("delete"),
		DetailedListFlag: viper.GetBool("detailed"),
		FileSize:         viper.GetString("size"),
		FileType:         fileType,
		OperatorType:     operatorType,
		Tolerance:        viper.GetFloat64("tolerance"),
		Results:          make(map[string][]string),
	}

	Run(finder)
}

// #endregion

// #region Main Logic
func main() {
	utils.ClearConsole()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Run(ff types.FileFinder) {
	fileSizeBytes, err := utils.ConvertSizeToBytes(ff.FileSize)

	if err != nil {
		pterm.Error.Printf("Error converting file size: %v\n", err)
		return
	}

	// Format the file size for logging
	fileSizeStr := utils.FormatSize(fileSizeBytes)

	// New file size based on tolerance
	toleranceSizeBytes, err := utils.CalculateToleranceBytes(ff.FileSize, ff.Tolerance)
	if err != nil {
		pterm.Error.Printf("Error calculating the tolerance size %v: %v\n", ff.Tolerance, err)
		return
	}

	pterm.Info.Printf("Searching for files of type %v %s %s...\n",
		ff.FileType,
		strings.ToLower(string(ff.OperatorType)),
		fileSizeStr,
	)

	utils.FindAndDisplay(ff, fileSizeBytes, toleranceSizeBytes, ff.DetailedListFlag)

	if ff.DeleteFlag {
		utils.DeleteFiles(ff)
	}
}

// #endregion
