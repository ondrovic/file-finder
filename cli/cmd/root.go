package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	sharedTypes "github.com/ondrovic/common/types"
	sharedUtils "github.com/ondrovic/common/utils"
	sharedFormatters "github.com/ondrovic/common/utils/formatters"

	"file-finder/internal/types"
	"file-finder/internal/utils"
	"file-finder/internal/utils/cli"
)

var (
	options     = types.CliFlags{}
	application sharedTypes.Application
	version     string

	rootCmd *cobra.Command
)

func initConfig() {
	viper.SetEnvPrefix(application.Name)
	viper.AutomaticEnv()
}

func init() {
	cobra.OnInitialize(initConfig)

	appName := "File-Finder"

	appNameToLower, err := sharedFormatters.ToLower(appName)
	if err != nil {
		// TODO: handle error
		return
	}

	application = sharedTypes.Application{
		Name:        appNameToLower,
		Description: "Cli to find and display files",
		Style:       sharedTypes.Styles{}, // TODO: think we are going to remove the style option
		Version:     sharedFormatters.GetVersion(version, "0.0.0-local-dev"),
	}

	rootCmd = &cobra.Command{
		Use:   fmt.Sprintf("%s <root-directory> [flags]", appNameToLower),
		Short: application.Description,
		Long:  application.Description,
		Args:  cobra.ExactArgs(1),
		RunE:  run,
	}

	rootCmd.SetVersionTemplate(`{{printf "Version: %s\n" .Version}}`)

	cli.RegisterBoolFlag(rootCmd, "display-app-banner", "b", false, "Whether or not to display the application banner", &options.DisplayApplicationBanner)
	cli.RegisterBoolFlag(rootCmd, "display-detailed-results", "d", false, "Display detailed results", &options.DisplayDetailedResults)
	// cli.RegisterBoolFlag(rootCmd, "list-duplicate-files", "u", false, "Lists duplicate files", &options.ListDuplicateFiles)
	cli.RegisterBoolFlag(rootCmd, "remove-files", "r", false, "Remove found files", &options.RemoveFiles)
	cli.RegisterFloat64Flag(rootCmd, "tolerance-size", "l", 0.05, "File size tolerance", &options.ToleranceSize)
	cli.RegisterStringFlag(rootCmd, "file-name-filter", "f", "", "Name to filter results by", &options.FileNameFilter)
	cli.RegisterStringFlag(rootCmd, "file-size-filter", "s", "", "File size to search for (1 KB, 1 MB, 1 GB)", &options.FileSizeFilter)
	cli.RegisterStringFlag(rootCmd, "file-type-filter", "t", string(sharedTypes.FileTypes.Any), "File type to search for (Any, Archive, Documents, Image, Video)", &options.FileTypeFilter)
	cli.RegisterStringFlag(rootCmd, "operator-type", "o", string(sharedTypes.OperatorTypes.EqualTo), "Operator to apply on file size\n(EqualTo: 'et', 'equal to', 'equal', '==')\n(GreaterThan: 'gt','greater', 'greater than', '>')\n(GreaterThanEqualTo: 'gte', 'greater than or equal to', 'greaterthanorequalto', '>=')\n(LessThan: 'lt', 'less', 'less than', 'lessthan', '<')\n(LessThanEqualTo: 'lte', 'less than or equal to',  'lessthanorequalto', '<='))", &options.OperatorTypeFilter)

	viper.BindPFlags(rootCmd.Flags())
}

func run(cmd *cobra.Command, args []string) error {

	fileTypeFilter := sharedUtils.ToFileType(viper.GetString("file-type-filter"))
	operatorType := sharedUtils.ToOperatorType(viper.GetString("operator-type"))

	removeFiles := viper.GetBool("remove-files")
	displayDetailedResults := viper.GetBool("display-detailed-results")

	if err := cli.ValidateInputs(fileTypeFilter, operatorType, removeFiles, displayDetailedResults); err != nil {
		return err
	}

	fileFinder := types.CliFlags{
		DisplayApplicationBanner: viper.GetBool("display-app-banner"),
		DisplayDetailedResults:   displayDetailedResults,
		FileNameFilter:           viper.GetString("file-name-filter"),
		FileSizeFilter:           viper.GetString("file-size-filter"),
		FileTypeFilter:           fileTypeFilter,
		// ListDuplicateFiles:       viper.GetBool("list-duplicate-files"),
		RemoveFiles:        removeFiles,
		ToleranceSize:      viper.GetFloat64("tolerance-size"),
		OperatorTypeFilter: operatorType,
		Results:            make(map[string][]string),
		RootDirectory:      args[0],
	}

	if err := findFiles(fileFinder); err != nil {
		return err
	}

	return nil
}

func findFiles(ff types.CliFlags) error {
	// TODO: fix types
	if files, err := utils.FindAndDisplayFiles(ff); err != nil {
		return err
	} else if ff.RemoveFiles {
		// TODO: add error handling
		utils.DeleteFiles(files)
	}

	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// TODO: handle error
		return
	}
}
