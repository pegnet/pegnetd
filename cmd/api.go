package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pegnet/pegnetd/node"

	"github.com/Factom-Asset-Tokens/fatd/fat"

	"github.com/pegnet/pegnetd/config"
	"github.com/spf13/viper"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnet/cmd"
	"github.com/pegnet/pegnetd/fat/fat2"
	"github.com/pegnet/pegnetd/srv"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(balances)
	rootCmd.AddCommand(balance)

	tx.Flags()
	rootCmd.AddCommand(tx)
}

var tx = &cobra.Command{
	Use:   "newtx <ECAddress> <SOURCE> <DESTINATION> <ASSET> <AMOUNT>",
	Short: "Builds and submits a pegnet transaction",
	Example: "pegnetd newtx EC3eX8VxGH64Xv3NFd9g4Y7PxSMnH3EGz5jQQrrQS8VZGnv4JY2K FA32xV6SoPBSbAZAVyuiHWwyoMYhnSyMmAHZfK29H8dx7bJXFLja" +
		" FA33kNzXwUt3cn4tLR56kyHEAryazAGPuMC6GjUubSbwrrNv8e7t PEG 200 ",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: cmd.CombineCobraArgs(
		cmd.CustomArgOrderValidationBuilder(
			true,
			cmd.ArgValidatorECAddress,
			cmd.ArgValidatorFCTAddress,
			cmd.ArgValidatorFCTAddress,
			cmd.ArgValidatorAsset,
			cmd.ArgValidatorFCTAmount),
	),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		cl := node.FactomClientFromConfig(viper.GetViper())
		payment, source, dest, asset, amt := args[0], args[1], args[2], args[3], args[4]
		// Build the transaction from the args
		var trans fat2.Transaction

		aType := fat2.StringToTicker(asset)
		if aType == fat2.PTickerInvalid {
			cmd.PrintErrf("invalid ticker type\n")
			os.Exit(1)
		}

		amount := FactoidToFactoshi(amt)
		if amount == -1 {
			cmd.PrintErrf("invalid amount specified\n")
			os.Exit(1)
		}

		// Set the input
		if trans.Input.Address, err = factom.NewFAAddress(source); err != nil {
			cmd.PrintErrf("failed to parse input: %s\n", err.Error())
			os.Exit(1)
		}

		// Get out private key
		priv, err := trans.Input.Address.GetFsAddress(cl)
		if err != nil {
			cmd.PrintErrf("unable to get private key : %s\n", err.Error())
			os.Exit(1)
		}

		trans.Input.Type = aType
		trans.Input.Amount = uint64(amount)

		pBals, err := queryBalances(source)
		if err != nil {
			cmd.PrintErrf("failed to get asset balance: %s", err.Error())
			os.Exit(1)
		}

		if pBals[aType] < trans.Input.Amount {
			cmd.PrintErrf("not enough %s to cover the transaction", aType.String())
			os.Exit(1)
		}

		// Set the output
		trans.Transfers = make([]fat2.AddressAmountTuple, 1)
		trans.Transfers[0].Amount = uint64(amount)
		if trans.Transfers[0].Address, err = factom.NewFAAddress(dest); err != nil {
			cmd.PrintErrf("failed to parse input: %s\n", err.Error())
			os.Exit(1)
		}

		// Sign the tx and make an entry
		var e fat.Entry
		content, err := json.Marshal(trans)
		if err != nil {
			cmd.PrintErrf("failed to marshal tx: %s", err.Error())
			os.Exit(1)
		}

		e.Content = content
		e.ChainID = &node.TransactionChain
		e.Sign(priv)

		ec, err := factom.NewECAddress(payment)
		if err != nil {
			cmd.PrintErrf("failed to parse input: %s\n", err.Error())
			os.Exit(1)
		}

		bal, err := ec.GetBalance(cl)
		if err != nil {
			cmd.PrintErrf("failed to get ec balance: %s\n", err.Error())
			os.Exit(1)
		}

		if cost, err := e.Cost(false); err != nil || uint64(cost) > bal {
			cmd.PrintErrf("not enough ec balance for the transaction")
			os.Exit(1)
		}

		es, err := ec.GetEsAddress(cl)
		if err != nil {
			cmd.PrintErrf("failed to parse input: %s\n", err.Error())
			os.Exit(1)
		}

		txid, err := e.ComposeCreate(cl, es, false)
		if err != nil {
			cmd.PrintErrf("failed to submit entry: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Printf("transaction sent : %s\n", txid)
	},
}

var balance = &cobra.Command{
	Use:              "balance",
	Short:            "Fetch the balance for a given asset and address",
	Example:          "pegnetd balance PEG FA2CEc2JSkhuckEXy42K111MvM9bycUDkbrrHjd9bNkBfvPBSGKd",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: cmd.CombineCobraArgs(
		cmd.CustomArgOrderValidationBuilder(false, cmd.ArgValidatorAssetAndAll, cmd.ArgValidatorFCTAddress),
		cobra.MinimumNArgs(1)),
	Run: func(cmd *cobra.Command, args []string) {
		res, err := queryBalances(args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(res[fat2.StringToTicker(args[0])])
	},
}

var balances = &cobra.Command{
	Use:              "balances",
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

		data, err := json.Marshal(res)
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
