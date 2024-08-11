package main

import (
	// "flag"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"file-finder/internal/types"
	"file-finder/internal/utils"

	commonTypes "github.com/ondrovic/common/types"
	commonUtils "github.com/ondrovic/common/utils"
)

// #region Cli Setup
var (
	rootCmd = &cobra.Command{
		Use:   "file-finder [root-directory]",
		Short: "Find files of specified size and type",
		Args:  cobra.ExactArgs(1),
		Run:   run,
	}

	options = types.FileFinder{}
)

func registerBoolFlag(cmd *cobra.Command, name, shorthand string, value bool, usage string, target *bool) {
	if !value {
		usage = usage + "\n (default false)"
	} else {
		usage = usage + "\n"
	}
	cmd.Flags().BoolVarP(target, name, shorthand, value, usage)
}

func registerStringFlag(cmd *cobra.Command, name, shorthand, value, usage string, target interface{}, completionFunc func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)) {
	targetValue := reflect.ValueOf(target)

	// Ensure target is a pointer
	if targetValue.Kind() != reflect.Ptr {
		panic("target must be a pointer")
	}

	// // Dereference the pointer to get the actual value
	elemValue := targetValue.Elem()

	// Use the StringVarP function with a temporary string variable
	var tempValue string
	cmd.Flags().StringVarP(&tempValue, name, shorthand, value, usage+"\n")

	// Update the original target with the value from tempValue
	elemValue.SetString(tempValue)

	if completionFunc != nil {
		cmd.RegisterFlagCompletionFunc(name, completionFunc)
	}
}

func registerFloat64Flag(cmd *cobra.Command, name, shorthand string, value float64, usage string, target *float64) {
	cmd.Flags().Float64VarP(target, name, shorthand, value, usage+"\n")
}

func bindFlags(cmd *cobra.Command) {
    cmd.Flags().VisitAll(func(f *pflag.Flag) {
        viper.BindPFlag(f.Name, f)
    })
}

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long:  "To load completions: ...", // Simplified for brevity
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	registerBoolFlag(rootCmd, "delete", "r", false, "Delete found files", &options.DeleteFlag)
	registerBoolFlag(rootCmd, "detailed", "d", false, "Display detailed results", &options.DetailedListFlag)
	registerFloat64Flag(rootCmd, "tolerance", "l", 0.05, "File size tolerance", &options.Tolerance)
	registerStringFlag(rootCmd, "type", "t", string(commonTypes.FileTypes.Video), "File type to search for (Any, Archive, Documents, Image, Video)", &options.FileType, nil)
	registerStringFlag(rootCmd, "operator", "o", string(commonTypes.OperatorTypes.EqualTo), "Operator to apply on file size\n(EqualTo: 'et', 'equal to', 'equal', '==')\n(GreaterThan: 'gt','greater', 'greater than', '>')\n(GreaterThanEqualTo: 'gte', 'greater than or equal to', 'greaterthanorequalto', '>=')\n(LessThan: 'lt', 'less', 'less than', 'lessthan', '<')\n(LessThanEqualTo: 'lte', 'less than or equal to',  'lessthanorequalto', '<='))", &options.OperatorType, nil)
	registerStringFlag(rootCmd, "size", "s", "", "File size to search for (1 KB, 1 MB, 1 GB)", &options.FileSize, nil)

	rootCmd.AddCommand(newCompletionCmd())
	
	// Bind flags with viper
	bindFlags(rootCmd)
}

func initConfig() {
	viper.SetEnvPrefix("FF")
	viper.AutomaticEnv()
}

func run(cmd *cobra.Command, args []string) {
	
	fileType := commonUtils.ToFileType(string(viper.GetString("type")))
	operatorType := commonUtils.ToOperatorType(string(viper.GetString("operator")))

	if fileType == "" {
		pterm.Error.Printf("invalid file type: %s", viper.GetString("type"))
	}
	
	if operatorType == "" {
		pterm.Error.Printf("invalid operator type: %s", viper.GetString("operator"))
	}

	fileFinder := types.FileFinder{
		RootDir:          args[0],
		DeleteFlag:       viper.GetBool("delete"),
		DetailedListFlag: viper.GetBool("detailed"),
		FileSize:         viper.GetString("size"),
		FileType:         fileType,
		OperatorType:     operatorType,
		Tolerance:        viper.GetFloat64("tolerance"),
		Results:          make(map[string][]string),
	}

	Run(fileFinder)
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

	files, err := utils.FindAndDisplayFiles(ff, fileSizeBytes, ff.Tolerance, ff.DetailedListFlag)

	if err != nil {
		pterm.Error.Printf("Error calculating tolerances: %v\n", err)
		return
	}

	// fmt.Println(files)

	if ff.DeleteFlag {
		utils.DeleteFiles(files)
	}
}

// #endregion
