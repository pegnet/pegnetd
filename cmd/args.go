package cmd

import (
	"fmt"
	"strings"

	"github.com/pegnet/pegnet/common"
	"github.com/spf13/cobra"
)

// ArgValidatorAssetAndAll checks for valid asset or 'all'
func ArgValidatorAssetOrP(cmd *cobra.Command, arg string) error {
	list := common.AllAssets
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
