package cmd

import (
	"fmt"
	"strings"

	"github.com/Factom-Asset-Tokens/factom"

	"github.com/pegnet/pegnet/modules/opr"
	"github.com/spf13/cobra"
)

//
// These all come from Pegnet. We copy these functions vs import to not have to
// import the factom libraries
//

// CombineCobraArgs allows the combination of multiple PositionalArgs
func CombineCobraArgs(funcs ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range funcs {
			err := f(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// CustomArgOrderValidationBuilder return an arg validator. The arg validator
// will validate cli arguments based on the validation functions provided
// in the order of the validation functions.
//		Params:
//			strict		Enforce the number of args == number of validation funcs
//			valids		Validation functions
func CustomArgOrderValidationBuilder(strict bool, valids ...func(cmd *cobra.Command, args string) error) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if strict && len(valids) != len(args) {
			return fmt.Errorf("accepts %d arg(s), received %d", len(valids), len(args))
		}

		for i, arg := range args {
			if err := valids[i](cmd, arg); err != nil {
				return err
			}
		}
		return nil
	}
}

// ArgValidatorAssetAndAll checks for valid asset or 'all'
func ArgValidatorAssetOrP(cmd *cobra.Command, arg string) error {
	list := opr.V2Assets
	for _, an := range list {
		if strings.ToLower(arg) == strings.ToLower(an) {
			return nil
		}
		if strings.ToLower(arg) == strings.ToLower("p"+an) {
			return nil
		}
	}

	errstr := fmt.Sprintf("not a valid asset. Options include: %v", list)

	if strings.Contains(arg, "BTC") {
		errstr += "\nI see you put in 'BTC', did you mean 'pXBT'?"
	}
	if strings.Contains(arg, "BCH") {
		errstr += "\nI see you put in 'BCH', did you mean 'pXBC'?"
	}
	return fmt.Errorf(errstr)
}

func ArgValidatorFCTAmount(cmd *cobra.Command, arg string) error {
	// The FCT amount must not be beyond 1e8 divisible
	_, err := FactoidToFactoshi(arg)
	return err
}

// ArgValidatorECAddress checks for EC address
func ArgValidatorECAddress(cmd *cobra.Command, arg string) error {
	if len(arg) > 2 && arg[:2] != "EC" {
		return fmt.Errorf("EC addresses start with EC")
	}

	_, err := factom.NewECAddress(arg)
	if err != nil {
		return fmt.Errorf("%s is not a valid EC address: %s", arg, err.Error())
	}
	return nil
}

// ArgValidatorFCTAddress checks for FCT address
func ArgValidatorFCTAddress(cmd *cobra.Command, arg string) error {
	if len(arg) > 2 && arg[:2] != "FA" {
		return fmt.Errorf("FCT addresses start with FA")
	}

	_, err := factom.NewFAAddress(arg)
	if err != nil {
		return fmt.Errorf("%s is not a valid FCT address: %s", arg, err.Error())
	}
	return nil
}
