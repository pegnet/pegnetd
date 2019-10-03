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
	return fmt.Errorf("not a valid asset. Options include: %v", list)
}
