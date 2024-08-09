package main

import (
	"os"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"file-finder/internal/types"
	"file-finder/internal/utils"

	commonTypes "github.com/ondrovic/common/types"
	commonUtils "github.com/ondrovic/common/utils"
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

	rootCmd.Flags().StringP("type", "t", string(commonTypes.FileTypes.Video), "File type to search for (Any, Video, Image, Archive, Documents)\n")
	rootCmd.Flags().StringP("operator", "o", string(commonTypes.OperatorTypes.EqualTo), "Operator to apply on file size (Equal To, Greater Than, Greater Than Or Equal To, Less Than, Less Than Or Equal To)\n")
	rootCmd.Flags().BoolP("delete", "r", false, "Delete found files\n(default: false)")
	rootCmd.Flags().BoolP("detailed", "d", false, "Display detailed results\n(default: false)")
	rootCmd.Flags().StringP("size", "s", "", "File size to search for (1 KB, 1 MB, 1 GB)\n")
	rootCmd.Flags().Float64P("tolerance", "l", 0.05, "File size tolerance\n")

	rootCmd.MarkFlagRequired("size")

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
	fileType := commonUtils.ToFileType(viper.GetString("type"))
	operatorType := commonUtils.ToOperatorType(viper.GetString("operator"))

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
	commonUtils.ClearTerminalScreen(runtime.GOOS)
	if err := rootCmd.Execute(); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
}

func Run(ff types.FileFinder) {
	fileSizeBytes, err := commonUtils.ConvertStringSizeToBytes(ff.FileSize)

	if err != nil {
		pterm.Error.Printf("Error converting file size: %v\n", err)
		return
	}

	// Format the file size for logging
	fileSizeStr := commonUtils.FormatSize(fileSizeBytes)
	results, err := commonUtils.CalculateTolerances(fileSizeBytes, ff.Tolerance)

	if err != nil {
		pterm.Error.Printf("Error calculating tolerances: %v\n", err)
		return
	}

	// Calculate the tolerance size string
	toleranceSizeStr := ""
	if fileSizeStr != commonUtils.FormatSize(results.LowerBoundSize) || fileSizeStr != commonUtils.FormatSize(results.UpperBoundSize) {
		toleranceSizeStr = "( with a tolerance size of " + commonUtils.FormatSize(results.LowerBoundSize) + " and " + commonUtils.FormatSize(results.UpperBoundSize) + " )"
	}

	pterm.Info.Printf("Searching for files of type %v %s %s %s...\n",
		ff.FileType,
		strings.ToLower(string(ff.OperatorType)),
		fileSizeStr,
		toleranceSizeStr,
	)

	utils.FindAndDisplayFiles(ff, fileSizeBytes, ff.Tolerance, ff.DetailedListFlag)

	if ff.DeleteFlag {
		utils.DeleteFiles(ff)
	}
}

// #endregion
