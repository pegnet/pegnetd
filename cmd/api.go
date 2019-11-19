package cmd

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pegnet/pegnet/modules/conversions"

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

	get.AddCommand(getTX)
	get.AddCommand(getRates)
	getSpread.Flags().Bool("tol", true, "Use tolerances for spread calculation")
	get.AddCommand(getSpread)
	getTXs.Flags().Bool("burn", false, "Show burns")
	getTXs.Flags().Bool("cvt", false, "Show converions")
	getTXs.Flags().Bool("tran", false, "Show transfers")
	getTXs.Flags().Bool("coin", false, "Show coinbases")

	get.AddCommand(getTXs)
	rootCmd.AddCommand(get)

	//tx.Flags()
	rootCmd.AddCommand(tx)
	rootCmd.AddCommand(conv)

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

		priv, err := addr.GetFsAddress(cl)
		if err != nil {
			cmd.PrintErrf("unable to get private key: %s\n", err.Error())
			os.Exit(1)
		}

		rcd, _, err := factom.DecodeRCD(priv.RCD())
		if err != nil {
			cmd.PrintErrf("unable to decode private key: %s\n", err.Error())
			os.Exit(1)
		}

		balance, err := addr.GetBalance(cl)
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
		trans.Timestamp = time.Now()
		trans.InputCount = 1
		trans.ECOutputCount = 1
		trans.FCTInputs = append(trans.FCTInputs, factom.FactoidTransactionIO{
			Amount:  uint64(amount),
			Address: &faddr,
		})
		trans.ECOutputs = append(trans.ECOutputs, factom.FactoidTransactionIO{
			Amount:  0,
			Address: &fBurnAddress,
		})

		// the library requires at least one signature to "be populated"
		// fill in below with real sig
		trans.Signatures = append(trans.Signatures, factom.FactoidTransactionSignature{
			ReedeemCondition: rcd,
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

		err = cl.FactomdRequest("factoid-submit", params, &result)
		if err != nil {
			cmd.PrintErrf("unable to submit transaction: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println(result.Message)
		fmt.Printf("Transaction ID: %s\n", result.TXID)
	},
}

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
			ArgValidatorFCTAddress,
			ArgValidatorAssetOrP,
			ArgValidatorFCTAmount,
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
	Args: CombineCobraArgs(
		CustomArgOrderValidationBuilder(
			true,
			ArgValidatorECAddress,
			ArgValidatorFCTAddress,
			ArgValidatorAssetOrP,
			ArgValidatorFCTAmount,
			ArgValidatorFCTAddress),
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
	Args: CombineCobraArgs(
		CustomArgOrderValidationBuilder(true, ArgValidatorAssetOrP, ArgValidatorFCTAddress),
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
	Args: CombineCobraArgs(
		CustomArgOrderValidationBuilder(true, ArgValidatorFCTAddress),
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

func queryBalances(humanAddress string) (srv.ResultPegnetTickerRateMap, error) {
	cl := srv.NewClient()
	cl.PegnetdServer = viper.GetString(config.Pegnetd)
	addr, err := factom.NewFAAddress(humanAddress)
	if err != nil {
		// TODO: Better error
		fmt.Println("1", err)
		os.Exit(1)
	}

	var res srv.ResultPegnetTickerRateMap
	err = cl.Request("get-pegnet-balances", srv.ParamsGetPegnetBalances{&addr}, &res)
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
		" provided will be displayed.",
	Example:          "pegnetd txs 07cebdd5d3f5216f36f792d71f030af07ddaa99147929d9af477833ee4c586a5",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args:             cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var height int
		// determine the params
		var params srv.ParamsGetPegnetTransaction

		// An entryhash?
		bytes, err := hex.DecodeString(args[0])
		if err == nil && len(bytes) == 32 {
			params.Hash = args[0]
			goto FoundParams
		}

		// A factoid address maybe?
		_, err = factom.NewFAAddress(args[0])
		if err == nil {
			params.Address = args[0]
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
		var res srv.ResultPegnetTickerRateMap
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

var getSpread = &cobra.Command{
	Use:              "spread <height> <src-asset> <optional dst-asset>",
	Example:          "pegnetd get spread 219000 pFCT\npegnetd get spread 219000 pFCT pXBT",
	Short:            "Fetch the spread amount and percent for a trading pair",
	PersistentPreRun: always,
	PreRun:           SoftReadConfig,
	Args:             cobra.RangeArgs(2, 3),
	Run: func(cmd *cobra.Command, args []string) {
		height, err := strconv.Atoi(args[0])
		if height <= 0 || err != nil {
			cmd.PrintErrf("height must be a number greater than 0")
			os.Exit(1)
		}

		cl := srv.NewClient()
		cl.PegnetdServer = viper.GetString(config.Pegnetd)
		var res srv.ResultPegnetTickerQuoteMap
		uH := uint32(height)
		err = cl.Request("get-pegnet-spreads", srv.ParamsGetPegnetRates{Height: &uH}, &res)
		if err != nil {
			fmt.Printf("Failed to make RPC request\nDetails:\n%v\n", err)
			os.Exit(1)
		}

		src := fat2.StringToTicker(toP(args[1]))
		dst := fat2.PTickerUSD
		if len(args) == 3 {
			dst = fat2.StringToTicker(toP(args[2]))
		}

		if src == fat2.PTickerInvalid || dst == fat2.PTickerInvalid {
			fmt.Printf("Arguments must be valid tickers\n")
			os.Exit(1)
		}

		mkC, _ := conversions.Convert(1e8, res[src].MarketRate, res[dst].MarketRate)
		pgC, _ := conversions.Convert(1e8, res[src].MinTolerance(), res[dst].MaxTolerance())
		var _ = pgC
		pair := res[src].MakeBase(res[dst])

		fmt.Printf("Trading pair %s/%s\n", src, dst)
		format := "%20s: %s\n"

		spreadString := func(s pegnet.QuotePair) []interface{} {
			if s.Spread() == 0 {
				s = s.Flip()
			}
			if tol, _ := cmd.Flags().GetBool("tol"); tol {
				return []interface{}{"Tolerance Spread", FactoshiToFactoid(s.SpreadWithTolerance())}
			}
			return []interface{}{"Raw Spread", FactoshiToFactoid(s.Spread())}
		}

		spreadPString := func(s pegnet.QuotePair, base int64) []interface{} {
			if s.Spread() == 0 {
				s = s.Flip()
			}
			if tol, _ := cmd.Flags().GetBool("tol"); tol {
				return []interface{}{"Tolerance Spread %", fmt.Sprintf("%.4f", 100.0*float64(s.SpreadWithTolerance())/float64(base))}
			}
			return []interface{}{"Raw Spread %", fmt.Sprintf("%.4f", 100.0*float64(s.Spread())/float64(base))}
		}

		var _, _ = spreadPString, spreadString

		fmt.Printf(format, "Market Rate", FactoshiToFactoid(mkC))
		fmt.Printf(format, "Buy Price", FactoshiToFactoid(pair.BuyRate()))
		fmt.Printf(format, "Sell Price", FactoshiToFactoid(pair.SellRate()))
		//
		//fmt.Printf(format, spreadString(pair)...)
		//fmt.Printf(format, spreadPString(pair, mkC)...)

		fmt.Println()
		if dst != fat2.PTickerUSD {
			srcUsdPair := res[src].MakeBase(res[fat2.PTickerUSD])
			fmt.Println("pUSD Prices")
			fmt.Printf("%4s/pUSD:\n", src)
			fmt.Printf("\t"+format, "Market Price", FactoshiToFactoid(int64(res[src].MarketRate)))
			fmt.Printf("\t"+format, "Moving Avg Price", FactoshiToFactoid(int64(res[src].MovingAverage)))
			fmt.Printf("\t"+format, "Buy Price", FactoshiToFactoid(srcUsdPair.BuyRate()))
			fmt.Printf("\t"+format, "Sell Price", FactoshiToFactoid(srcUsdPair.SellRate()))

			dstUsdPair := res[dst].MakeBase(res[fat2.PTickerUSD])
			fmt.Printf("%4s/pUSD:\n", dst)
			fmt.Printf("\t"+format, "Market Price", FactoshiToFactoid(int64(res[dst].MarketRate)))
			fmt.Printf("\t"+format, "Moving Avg Price", FactoshiToFactoid(int64(res[dst].MovingAverage)))
			fmt.Printf("\t"+format, "Buy Price", FactoshiToFactoid(dstUsdPair.BuyRate()))
			fmt.Printf("\t"+format, "Sell Price", FactoshiToFactoid(dstUsdPair.SellRate()))
		}
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
