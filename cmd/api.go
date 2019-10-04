package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pegnet/pegnetd/node"

	"github.com/pegnet/pegnetd/config"
	"github.com/spf13/viper"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnet/cmd"
	"github.com/pegnet/pegnetd/fat/fat2"
	"github.com/pegnet/pegnetd/srv"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(balance)
	rootCmd.AddCommand(balances)
	rootCmd.AddCommand(status)

	//tx.Flags()
	rootCmd.AddCommand(tx)
	rootCmd.AddCommand(conv)
}

var conv = &cobra.Command{
	Use:     "newcvt <ECAddress> <FA-SOURCE> <SRC-ASSET> <AMOUNT> <DEST-ASSET>",
	Aliases: []string{"newconversion", "newconvert"},
	Short:   "Builds and submits a pegnet conversion",
	Example: "pegnetd newcvt EC3eX8VxGH64Xv3NFd9g4Y7PxSMnH3EGz5jQQrrQS8VZGnv4JY2K FA32xV6SoPBSbAZAVyuiHWwyoMYhnSyMmAHZfK29H8dx7bJXFLja" +
		" pFCT 100 pUSD ",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: cmd.CombineCobraArgs(
		cmd.CustomArgOrderValidationBuilder(
			true,
			cmd.ArgValidatorECAddress,
			cmd.ArgValidatorFCTAddress,
			ArgValidatorAssetOrP,
			cmd.ArgValidatorFCTAmount,
			ArgValidatorAssetOrP),
	),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		cl := node.FactomClientFromConfig(viper.GetViper())
		payment, source, srcAsset, amt, destAsset := args[0], args[1], args[2], args[3], args[4]

		// Build the transaction from the args
		var trans fat2.Transaction
		if err := setTransactionInput(&trans, cl, source, srcAsset, amt); err != nil {
			cmd.PrintErrln(err.Error())
			os.Exit(1)
		}

		if trans.Conversion, err = ticker(destAsset); err != nil {
			cmd.PrintErrln("invalid ticker type")
			os.Exit(1)
		}

		err, commit, reveal := signAndSend(&trans, cl, payment)
		if err != nil {
			cmd.PrintErrln(err.Error())
			os.Exit(1)
		}

		fmt.Printf("conversion sent:\n")
		fmt.Printf("\t%10s: %s\n", "EntryHash", reveal)
		fmt.Printf("\t%10s: %s\n", "Commit", commit)
	},
}

var tx = &cobra.Command{
	Use:   "newtx <ECAddress> <FA-SOURCE> <ASSET> <AMOUNT> <FA-DESTINATION>",
	Short: "Builds and submits a pegnet transaction",
	Example: "pegnetd newtx EC3eX8VxGH64Xv3NFd9g4Y7PxSMnH3EGz5jQQrrQS8VZGnv4JY2K " +
		" FA33kNzXwUt3cn4tLR56kyHEAryazAGPuMC6GjUubSbwrrNv8e7t PEG 200 FA32xV6SoPBSbAZAVyuiHWwyoMYhnSyMmAHZfK29H8dx7bJXFLja",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: cmd.CombineCobraArgs(
		cmd.CustomArgOrderValidationBuilder(
			true,
			cmd.ArgValidatorECAddress,
			cmd.ArgValidatorFCTAddress,
			ArgValidatorAssetOrP,
			cmd.ArgValidatorFCTAmount,
			cmd.ArgValidatorFCTAddress),
	),
	Run: func(cmd *cobra.Command, args []string) {
		cl := node.FactomClientFromConfig(viper.GetViper())
		payment, source, asset, amt, dest := args[0], args[1], args[2], args[3], args[4]

		// Build the transaction from the args
		var trans fat2.Transaction
		if err := setTransactionInput(&trans, cl, source, asset, amt); err != nil {
			cmd.PrintErrln(err.Error())
			os.Exit(1)
		}

		if err := setTransferOutput(&trans, cl, dest, amt); err != nil {
			cmd.PrintErrln(err.Error())
			os.Exit(1)
		}

		err, commit, reveal := signAndSend(&trans, cl, payment)
		if err != nil {
			cmd.PrintErrln(err.Error())
			os.Exit(1)
		}

		fmt.Printf("transaction sent:\n")
		fmt.Printf("\t%10s: %s\n", "EntryHash", reveal)
		fmt.Printf("\t%10s: %s\n", "Commit", commit)
	},
}

var balance = &cobra.Command{
	Use:              "balance <asset> <factoid-address>",
	Short:            "Fetch the balance for a given asset and address",
	Example:          "pegnetd balance PEG FA2CEc2JSkhuckEXy42K111MvM9bycUDkbrrHjd9bNkBfvPBSGKd",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: cmd.CombineCobraArgs(
		cmd.CustomArgOrderValidationBuilder(false, ArgValidatorAssetOrP, cmd.ArgValidatorFCTAddress),
		cobra.MinimumNArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		res, err := queryBalances(args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		ticker := fat2.StringToTicker(toP(args[0]))
		balance := res[ticker]
		humanBal := FactoshiToFactoid(int64(balance))
		fmt.Printf("%s %s\n", humanBal, ticker.String())
	},
}

var balances = &cobra.Command{
	Use:              "balances <factoid-address>",
	Short:            "Fetch all balances for a given factoid address",
	Example:          "pegnetd balances FA2CEc2JSkhuckEXy42K111MvM9bycUDkbrrHjd9bNkBfvPBSGKd",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: cmd.CombineCobraArgs(
		cmd.CustomArgOrderValidationBuilder(false, cmd.ArgValidatorFCTAddress),
		cobra.MinimumNArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		res, err := queryBalances(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Change the units to be human readable
		humanBals := make(map[string]string)
		for k, bal := range res {
			humanBals[k.String()] = FactoshiToFactoid(int64(bal))
		}

		data, err := json.Marshal(humanBals)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
	},
}

func queryBalances(humanAddress string) (srv.ResultGetPegnetBalances, error) {
	cl := srv.NewClient()
	cl.PegnetdServer = viper.GetString(config.Pegnetd)
	addr, err := factom.NewFAAddress(humanAddress)
	if err != nil {
		// TODO: Better error
		fmt.Println("1", err)
		os.Exit(1)
	}

	var res srv.ResultGetPegnetBalances
	err = cl.Request("get-pegnet-balances", srv.ParamsGetPegnetBalances{&addr}, &res)
	if err != nil {
		// TODO: Better error
		fmt.Println("2", err)
		os.Exit(1)
	}

	return res, nil
}

var status = &cobra.Command{
	Use:              "status",
	Short:            "Fetch the current sync status for the pegnetd node",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Run: func(cmd *cobra.Command, args []string) {
		cl := srv.NewClient()
		cl.PegnetdServer = viper.GetString(config.Pegnetd)
		var res srv.ResultGetSyncStatus
		err := cl.Request("get-sync-status", nil, &res)
		if err != nil {
			fmt.Printf("Failed to make RPC request\nDetails:\n%v\n", err)
			os.Exit(1)
		}

		data, err := json.Marshal(res)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
	},
}

func toP(asset string) string {
	if strings.ToLower(asset) == "PEG" {
		return "PEG"
	}

	if strings.ToLower(asset)[0] != 'p' {
		return "p" + strings.ToUpper(asset)
	}
	return asset
}
