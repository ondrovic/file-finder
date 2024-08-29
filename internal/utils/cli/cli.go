package cli

import (
	"fmt"
	"reflect"

	sharedTypes "github.com/ondrovic/common/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func RegisterBoolFlag(cmd *cobra.Command, name, shorthand string, value bool, usage string, target *bool) {
	if !value {
		usage = usage + "\n (default false)"
	} else {
		usage = usage + "\n"
	}
	cmd.Flags().BoolVarP(target, name, shorthand, value, usage)
}

func RegisterStringFlag(cmd *cobra.Command, name, shorthand, value, usage string, target interface{}) {
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
}

func RegisterFloat64Flag(cmd *cobra.Command, name, shorthand string, value float64, usage string, target *float64) {
	cmd.Flags().Float64VarP(target, name, shorthand, value, usage+"\n")
}

func ValidateInputs(fileTypeFilter sharedTypes.FileType, operatorType sharedTypes.OperatorType, removeFiles, displayDetailedResults bool) error {
	if fileTypeFilter == "" {
		return fmt.Errorf("invalid file type: %s", viper.GetString("file-type-filter"))
	}

	if operatorType == "" {
		return fmt.Errorf("invalid operator type: %s", viper.GetString("operator-type"))
	}

	if removeFiles && !displayDetailedResults {
		return fmt.Errorf("the flags --remove-files (-r) and --display-detailed-results (-d) must be used together, any other combination isn't supported")
	}

	return nil
}
