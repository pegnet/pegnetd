package cmd

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/config"
	"github.com/pegnet/pegnetd/fat/fat2"
	"github.com/pegnet/pegnetd/node"
	"github.com/pegnet/pegnetd/node/pegnet"
	"github.com/pegnet/pegnetd/srv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(balance)
	rootCmd.AddCommand(balances)
	rootCmd.AddCommand(issuance)
	rootCmd.AddCommand(status)
	rootCmd.AddCommand(burn)
	rich.Flags().Int("count", 100, "The top X address")
	rootCmd.AddCommand(rich)

	get.AddCommand(getTX)
	get.AddCommand(getRates)
	getTXs.Flags().Bool("burn", false, "Show burns")
	getTXs.Flags().Bool("cvt", false, "Show converions")
	getTXs.Flags().Bool("tran", false, "Show transfers")
	getTXs.Flags().Bool("coin", false, "Show coinbases")
	getTXs.Flags().String("asset", "", "Filter by specific asset")
	getTXs.Flags().Int("offset", 0, "Specify an offset for pagination")

	get.AddCommand(getTXs)
	rootCmd.AddCommand(get)

	//tx.Flags()
	rootCmd.AddCommand(tx)
	rootCmd.AddCommand(conv)

}

var rich = &cobra.Command{
	Use:              "richlist [ASSET]",
	Short:            "Get a list of richest addresses",
	Example:          "richlist PEG --count=1",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args:             cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cl := srv.NewClient()
		cl.PegnetdServer = viper.GetString(config.Pegnetd)
		count, _ := cmd.Flags().GetInt("count")
		if count == 0 {
			count = 100
		}

		if len(args) > 0 {
			ticker := fat2.StringToTicker(args[0])
			if ticker == fat2.PTickerInvalid {
				cmd.PrintErrln(fmt.Errorf("invalid asset specified"))
				os.Exit(1)
			}

			assetRich(cl, ticker.String(), count)
		} else {
			globalRich(cl, count)
		}
	},
}

func assetRich(cl *srv.Client, asset string, count int) {
	var params srv.ParamsGetRichList
	params.Asset = asset
	params.Count = count

	var res []srv.ResultGetRichList
	err := cl.Request("get-rich-list", params, &res)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Top %d %s Rich List\n", count, asset)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Pos\tAddress\t%s\tpUSD\t\n", asset)
	fmt.Fprintf(tw, "---\t-------\t%s\t----\t\n", strings.Repeat("-", len(asset)))
	for i, e := range res {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t\n", i+1, e.Address, FactoshiToFactoid(int64(e.Amount)), FactoshiToFactoid(int64(e.Equiv)))
	}
	tw.Flush()
}

func globalRich(cl *srv.Client, count int) {
	var params srv.ParamsGetGlobalRichList
	params.Count = count

	var res []srv.ResultGlobalRichList
	err := cl.Request("get-global-rich-list", params, &res)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Top %d Global Rich List\n", count)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Pos\tAddress\tpUSD\t\n")
	fmt.Fprintf(tw, "---\t-------\t----\t\n")
	for i, e := range res {
		fmt.Fprintf(tw, "%d\t%s\t%s\t\n", i+1, e.Address, FactoshiToFactoid(int64(e.Equiv)))
	}
	tw.Flush()
}

var burn = &cobra.Command{
	Use:              "burn <FA-SOURCE> <AMOUNT>",
	Short:            "Converts FCT into pFCT",
	Example:          "pegnetd burn FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q 50",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: CombineCobraArgs(
		CustomArgOrderValidationBuilder(
			true,
			ArgValidatorFCTAddress,
			ArgValidatorFCTAmount,
		),
	),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		cl := node.FactomClientFromConfig(viper.GetViper())
		source, amt := args[0], args[1]

		amount, err := FactoidToFactoshi(amt)
		if err != nil {
			cmd.PrintErrln(fmt.Errorf("invalid amount specified"))
			os.Exit(1)
		}

		addr, err := factom.NewFAAddress(source)
		if err != nil {
			cmd.PrintErrln("invalid input address specified")
			os.Exit(1)
		}
		faddr := factom.Bytes32(addr)

		priv, err := addr.GetFsAddress(nil, cl)
		if err != nil {
			cmd.PrintErrf("unable to get private key: %s\n", err.Error())
			os.Exit(1)
		}

		rcd, _, err := factom.DecodeRCD(priv.RCD())
		if err != nil {
			cmd.PrintErrf("unable to decode private key: %s\n", err.Error())
			os.Exit(1)
		}

		rcd1, ok := rcd.(*factom.RCD1)
		if !ok {
			cmd.PrintErrln("the address is not compatible with factoid transactions, must be rcd type 1")
			os.Exit(1)
		}

		balance, err := addr.GetBalance(nil, cl)
		if err != nil {
			cmd.PrintErrln("unable to retrieve balance:" + err.Error())
			os.Exit(1)
		}

		if balance < uint64(amount) {
			cmd.PrintErrf("not enough balance to cover the amount. balance = %s\n", FactoshiToFactoid(int64(balance)))
			os.Exit(1)
		}

		burnAddress, _ := factom.NewECAddress(node.BurnAddress)
		fBurnAddress := factom.Bytes32(burnAddress)

		var trans factom.FactoidTransaction
		trans.Version = 2
		trans.TimestampSalt = time.Now()
		trans.FCTInputs = append(trans.FCTInputs, factom.FactoidTransactionIO{
			Amount:  uint64(amount),
			Address: faddr,
		})
		trans.ECOutputs = append(trans.ECOutputs, factom.FactoidTransactionIO{
			Amount:  0,
			Address: fBurnAddress,
		})

		// the library requires at least one signature to "be populated"
		// fill in below with real sig
		trans.Signatures = append(trans.Signatures, factom.FactoidTransactionSignature{
			ReedeemCondition: *rcd1,
			SignatureBlock:   nil,
		})

		data, err := trans.MarshalLedgerBinary()
		if err != nil { // should not happen
			cmd.PrintErrf("unable to marshal for signature: %s\n", err.Error())
			os.Exit(1)
		}

		sig := ed25519.Sign(priv.PrivateKey(), data)
		trans.Signatures[0].SignatureBlock = sig

		raw, err := trans.MarshalBinary()
		if err != nil { // should not happen
			cmd.PrintErrf("unable to marshal transaction: %s\n", err.Error())
			os.Exit(1)
		}

		params := struct {
			Hex string `json:"transaction"`
		}{Hex: fmt.Sprintf("%x", raw)}

		var result struct {
			Message string `json:"message"`
			TXID    string `json:"txid"`
		}

		err = cl.FactomdRequest(nil, "factoid-submit", params, &result)
		if err != nil {
			cmd.PrintErrf("unable to submit transaction: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println(result.Message)
		fmt.Printf("Transaction ID: %s\n", result.TXID)
	},
}

var outputFEWarning = "The address you are sending to is an Ethereum linked address. In transactions, the output address will be displayed as %s."

var conv = &cobra.Command{
	Use:     "newcvt <ECAddress> <FA-SOURCE> <SRC-ASSET> <AMOUNT> <DEST-ASSET>",
	Aliases: []string{"newconversion", "newconvert"},
	Short:   "Builds and submits a pegnet conversion",
	Example: "pegnetd newcvt EC3eX8VxGH64Xv3NFd9g4Y7PxSMnH3EGz5jQQrrQS8VZGnv4JY2K FA32xV6SoPBSbAZAVyuiHWwyoMYhnSyMmAHZfK29H8dx7bJXFLja" +
		" pFCT 100 pUSD ",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: CombineCobraArgs(
		CustomArgOrderValidationBuilder(
			true,
			ArgValidatorECAddress,
			ArgValidatorAddress(ADD_FA|ADD_FE|ADD_Fe),
			ArgValidatorAssetOrP,
			ArgValidatorFCTAmount,
			ArgValidatorAssetOrP),
	),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		cl := node.FactomClientFromConfig(viper.GetViper())
		payment, originalSource, srcAsset, amt, destAsset := args[0], args[1], args[2], args[3], args[4]

		// Let's check the pXXX -> pFCT first
		status := getStatus()
		if (destAsset == "pFCT" || destAsset == "FCT") && uint32(status.Current) >= node.OneWaypFCTConversions {
			cmd.PrintErrln(fmt.Sprintf("pXXX -> pFCT conversions are not allowed since block height %d. If you need to acquire pFCT, you have to burn FCT -> pFCT", node.OneWaypFCTConversions))
			os.Exit(1)
		}

		// Build the transaction from the args
		var trans fat2.Transaction
		if err := setTransactionInput(&trans, cl, originalSource, srcAsset, amt); err != nil {
			cmd.PrintErrln(err.Error())
			os.Exit(1)
		}

		if trans.Conversion, err = ticker(destAsset); err != nil {
			cmd.PrintErrln("invalid ticker type")
			os.Exit(1)
		}

		err, commit, reveal := signAndSend(originalSource, &trans, cl, payment)
		if err != nil {
			cmd.PrintErrln(err.Error())
			os.Exit(1)
		}

		fmt.Printf("conversion sent:\n")
		fmt.Printf("\t%10s: %s\n", "EntryHash", reveal)
		fmt.Printf("\t%10s: %s\n", "Commit", commit)
		printFeWarning(cmd, originalSource)
	},
}

var tx = &cobra.Command{
	Use:   "newtx <ECAddress> <FA-SOURCE> <ASSET> <AMOUNT> <FA-DESTINATION>",
	Short: "Builds and submits a pegnet transaction",
	Example: "pegnetd newtx EC3eX8VxGH64Xv3NFd9g4Y7PxSMnH3EGz5jQQrrQS8VZGnv4JY2K " +
		" FA33kNzXwUt3cn4tLR56kyHEAryazAGPuMC6GjUubSbwrrNv8e7t PEG 200 FA32xV6SoPBSbAZAVyuiHWwyoMYhnSyMmAHZfK29H8dx7bJXFLja",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: CombineCobraArgs(
		CustomArgOrderValidationBuilder(
			true,
			ArgValidatorECAddress,
			ArgValidatorAddress(ADD_FA|ADD_FE|ADD_Fe),
			ArgValidatorAssetOrP,
			ArgValidatorFCTAmount,
			ArgValidatorAddress(ADD_FA|ADD_FE|ADD_Fe)),
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

		// Before we sign and send, check the in/out rules
		err := addressRules(source, dest)
		if err != nil {
			cmd.PrintErrln(err.Error())
			os.Exit(1)
		}

		err, commit, reveal := signAndSend(source, &trans, cl, payment)
		if err != nil {
			cmd.PrintErrln(err.Error())
			os.Exit(1)
		}

		fmt.Printf("transaction sent:\n")
		fmt.Printf("\t%10s: %s\n", "EntryHash", reveal)
		fmt.Printf("\t%10s: %s\n", "Commit", commit)

		printFeWarning(cmd, source, dest)

		//printFeWarning(cmd, source, false,
		//	fmt.Sprintf("The address you are sending from is an Ethereum linked address. In transactions, the input address will be displayed as %%s. "+
		//		"Continue to use '%s'! DO NOT USE THIS FA ADDRESS DIRECTLY. LOSS OF FUNDS MAY RESULT!", source))
		//printFeWarning(cmd, dest, false,
		//	"The address you are sending to is an Ethereum linked address. In transactions, the output address will be displayed as %s.")
	},
}

var balance = &cobra.Command{
	Use:              "balance <asset> <factoid-address>",
	Short:            "Fetch the balance for a given asset and address",
	Example:          "pegnetd balance PEG FA2CEc2JSkhuckEXy42K111MvM9bycUDkbrrHjd9bNkBfvPBSGKd",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: CombineCobraArgs(
		CustomArgOrderValidationBuilder(true, ArgValidatorAssetOrP, ArgValidatorAddress(ADD_FA|ADD_FE|ADD_Fe)),
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
		printFeWarning(cmd, args[1])
	},
}

var balances = &cobra.Command{
	Use:              "balances <factoid-address>",
	Short:            "Fetch all balances for a given factoid address",
	Example:          "pegnetd balances FA2CEc2JSkhuckEXy42K111MvM9bycUDkbrrHjd9bNkBfvPBSGKd",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args: CombineCobraArgs(
		CustomArgOrderValidationBuilder(true, ArgValidatorAddress(ADD_FA|ADD_FE|ADD_Fe)),
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
		printFeWarning(cmd, args[0])
	},
}

func queryBalances(humanAddress string) (srv.ResultPegnetTickerMap, error) {
	cl := srv.NewClient()
	cl.PegnetdServer = viper.GetString(config.Pegnetd)
	addr, err := underlyingFA(humanAddress)
	if err != nil {
		// TODO: Better error
		fmt.Println("1", err)
		os.Exit(1)
	}

	var res srv.ResultPegnetTickerMap
	err = cl.Request("get-pegnet-balances", srv.ParamsGetPegnetBalances{addr.String()}, &res)
	if err != nil {
		// TODO: Better error
		fmt.Println("2", err)
		os.Exit(1)
	}

	return res, nil
}

var issuance = &cobra.Command{
	Use:              "issuance",
	Short:            "Fetch the current issuance of all assets",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Run: func(cmd *cobra.Command, args []string) {
		cl := srv.NewClient()
		cl.PegnetdServer = viper.GetString(config.Pegnetd)
		var res srv.ResultGetIssuance
		err := cl.Request("get-pegnet-issuance", nil, &res)
		if err != nil {
			fmt.Printf("Failed to make RPC request\nDetails:\n%v\n", err)
			os.Exit(1)
		}

		// Change the units to be human readable
		humanIssuance := make(map[string]string)
		for k, bal := range res.Issuance {
			humanIssuance[k.String()] = FactoshiToFactoid(int64(bal))
		}
		humanResult := struct {
			SyncStatus srv.ResultGetSyncStatus `json:"sync-status"`
			Issuance   map[string]string       `json:"issuance"`
		}{
			SyncStatus: res.SyncStatus,
			Issuance:   humanIssuance,
		}

		data, err := json.Marshal(humanResult)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
	},
}

var status = &cobra.Command{
	Use:              "status",
	Short:            "Fetch the current sync status for the pegnetd node",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Run: func(cmd *cobra.Command, args []string) {
		res := getStatus()
		data, err := json.Marshal(res)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
	},
}

func getProperties() srv.PegnetdProperties {
	cl := srv.NewClient()
	cl.PegnetdServer = viper.GetString(config.Pegnetd)
	var res srv.PegnetdProperties
	err := cl.Request("properties", nil, &res)
	if err != nil {
		return srv.PegnetdProperties{
			BuildVersion:  "Unknown/Unable",
			BuildCommit:   "Unknown/Unable",
			SQLiteVersion: "Unknown/Unable",
			GolangVersion: "Unknown/Unable",
		}
	}
	return res
}

func getStatus() srv.ResultGetSyncStatus {
	cl := srv.NewClient()
	cl.PegnetdServer = viper.GetString(config.Pegnetd)
	var res srv.ResultGetSyncStatus
	err := cl.Request("get-sync-status", nil, &res)
	if err != nil {
		fmt.Printf("Failed to make RPC request\nDetails:\n%v\n", err)
		os.Exit(1)
	}
	return res
}

var get = &cobra.Command{
	Use:   "get <subcommand>",
	Short: "Able to read pegnet related information from the daemon.",
}

var getTX = &cobra.Command{
	Use:              "tx <txid>",
	Short:            "Fetch the transaction by the given entryhash",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args:             cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, _, err := pegnet.SplitTxID(args[0])
		if err != nil {
			cmd.PrintErrf("txid is invalid: %s", err.Error())
		}

		cl := srv.NewClient()
		cl.PegnetdServer = viper.GetString(config.Pegnetd)
		var res srv.ResultGetTransactions
		err = cl.Request("get-transaction", srv.ParamsGetPegnetTransaction{TxID: args[0]}, &res)
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

var getTXs = &cobra.Command{
	Use:   "txs <entryhash | FA address | height>",
	Short: "Fetch all transactions for an entryhash, FA address, or height",
	Long: "Fetch all transactions for an entryhash, FA address, or height. " +
		"If a --burn, --cvt, --tran, or --coin is provided, then only the flags" +
		" provided will be displayed. If you specify --asset=pAsset, only transactions" +
		" involving that asset will be returned.",
	Example:          "pegnetd txs 07cebdd5d3f5216f36f792d71f030af07ddaa99147929d9af477833ee4c586a5",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args:             cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var height int
		// determine the params
		var params srv.ParamsGetPegnetTransaction
		var add factom.FAAddress

		// An entryhash?
		bytes, err := hex.DecodeString(args[0])
		if err == nil && len(bytes) == 32 {
			params.Hash = args[0]
			goto FoundParams
		}

		// A factoid address maybe?
		add, err = underlyingFA(args[0])
		if err == nil {
			// Place warning at the bottom
			defer printFeWarning(cmd, args[0])
			params.Address = add.String()
			goto FoundParams
		}

		// Ok, maybe it's a height!
		height, err = strconv.Atoi(args[0])
		if err == nil {
			params.Height = height
			goto FoundParams
		}

		// I give up.
		cmd.PrintErrf("param invalid. could not determine type")
	FoundParams:

		params.Conversion, _ = cmd.Flags().GetBool("cvt")
		params.Burn, _ = cmd.Flags().GetBool("burn")
		params.Transfer, _ = cmd.Flags().GetBool("tran")
		params.Coinbase, _ = cmd.Flags().GetBool("coin")
		params.Asset, _ = cmd.Flags().GetString("asset")
		params.Offset, _ = cmd.Flags().GetInt("offset")

		cl := srv.NewClient()
		cl.PegnetdServer = viper.GetString(config.Pegnetd)
		var res srv.ResultGetTransactions
		err = cl.Request("get-transactions", params, &res)
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

var getRates = &cobra.Command{
	Use:              "rates <height>",
	Short:            "Fetch the pegnet quotes for the assets at a given height (if their are quotes)",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args:             cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		height, err := strconv.Atoi(args[0])
		if height <= 0 || err != nil {
			cmd.PrintErrf("height must be a number greater than 0")
			os.Exit(1)
		}

		cl := srv.NewClient()
		cl.PegnetdServer = viper.GetString(config.Pegnetd)
		var res srv.ResultPegnetTickerMap
		uH := uint32(height)
		err = cl.Request("get-pegnet-rates", srv.ParamsGetPegnetRates{Height: &uH}, &res)
		if err != nil {
			fmt.Printf("Failed to make RPC request\nDetails:\n%v\n", err)
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

func toP(asset string) string {
	if strings.ToLower(asset) == "PEG" {
		return "PEG"
	}

	if strings.ToLower(asset)[0] != 'p' {
		return "p" + strings.ToUpper(asset)
	}
	return asset
}
