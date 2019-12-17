package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"
	"github.com/pegnet/pegnetd/node"
)

func addressRules(input string, output string) error {
	if len(input) < 2 || len(output) < 2 {
		return fmt.Errorf("input or output is too short to be an address")
	}

	switch input[:2] {
	case "FA":
		if output[:2] == "FA" || output[:2] == "Fe" {
			// This is ok
			return nil
		}
		if output[:2] == "FE" {
			return fmt.Errorf("FA addresses can only send pAssets to FA or Fe addresses. You cannot send pAssets to a gateway address, FE, from an FA address. You must use an address that has a linked ethereum address to do that")
		}
		return fmt.Errorf("FA addresses can only send pAssets to FA or Fe addresses. It seems what you are trying to do it not allowed")
	case "Fe":
		if output[:2] == "FA" || output[:2] == "Fe" || output[:2] == "FE" {
			// This is ok
			return nil
		}

		return fmt.Errorf("Fe addresses can only send pAssets to FA, Fe, or FE addresses. It seems what you are trying to do it not allowed")
	}
	return nil
}

func signAndSend(source string, tx *fat2.Transaction, cl *factom.Client, payment string) (err error, commit *factom.Bytes32, reveal *factom.Bytes32) {
	// Get out private key
	// If the source is an Fe/FE address, we use the eth secret
	var priv factom.RCDSigner

	switch source[:2] {
	case "FA":
		priv, err = tx.Input.Address.GetFsAddress(nil, cl)
		if err != nil {
			return fmt.Errorf("[FA] unable to get private key: %s\n", err.Error()), nil, nil
		}
	case "Fe":
		addr := factom.FeAddress(tx.Input.Address)
		priv, err = addr.GetEthSecret(nil, cl)
		if err != nil {
			return fmt.Errorf("[Fe] unable to get private key: %s\n", err.Error()), nil, nil
		}
	case "FE":
		addr := factom.FEGatewayAddress(tx.Input.Address)
		priv, err = addr.GetEthSecret(nil, cl)
		if err != nil {
			return fmt.Errorf("[FE] unable to get private key: %s\n", err.Error()), nil, nil
		}
	}

	var txBatch fat2.TransactionBatch
	txBatch.Version = 1
	txBatch.Transactions = []fat2.Transaction{*tx}
	txBatch.Entry.ChainID = &node.TransactionChain

	// Sign the tx and make an entry
	entry, err := txBatch.Sign(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal tx: %s", err.Error()), nil, nil
	}
	txBatch.Entry = entry

	if err := txBatch.Validate(-1); err != nil {
		return fmt.Errorf("invalid tx: %s", err.Error()), nil, nil
	}

	ec, err := factom.NewECAddress(payment)
	if err != nil {
		return fmt.Errorf("failed to parse input: %s\n", err.Error()), nil, nil
	}

	bal, err := ec.GetBalance(nil, cl)
	if err != nil {
		return fmt.Errorf("failed to get ec balance: %s\n", err.Error()), nil, nil
	}

	if cost, err := txBatch.Entry.Cost(); err != nil || uint64(cost) > bal {
		return fmt.Errorf("not enough ec balance for the transaction"), nil, nil
	}

	es, err := ec.GetEsAddress(nil, cl)
	if err != nil {
		return fmt.Errorf("failed to parse input: %s\n", err.Error()), nil, nil
	}

	txid, err := txBatch.Entry.ComposeCreate(nil, cl, es)
	if err != nil {
		return fmt.Errorf("failed to submit entry: %s\n", err.Error()), nil, nil
	}

	return nil, &txid, txBatch.Entry.Hash
}

func setTransferOutput(tx *fat2.Transaction, cl *factom.Client, dest, amt string) error {
	var err error
	amount, err := FactoidToFactoshi(amt)
	if err != nil {
		return fmt.Errorf("invalid amount specified: %s\n", err.Error())
	}

	tx.Transfers = make([]fat2.AddressAmountTuple, 1)
	tx.Transfers[0].Amount = uint64(amount)
	if tx.Transfers[0].Address, err = underlyingFA(dest); err != nil {
		return fmt.Errorf("failed to parse input: %s\n", err.Error())
	}

	return nil
}

func setTransactionInput(tx *fat2.Transaction, cl *factom.Client, source, asset, amt string) error {
	var err error
	if tx.Input.Type, err = ticker(asset); err != nil {
		return err
	}

	amount, err := FactoidToFactoshi(amt)
	if err != nil {
		return fmt.Errorf("invalid amount specified: %s\n", err.Error())
	}
	tx.Input.Amount = uint64(amount)

	// Set the input
	if tx.Input.Address, err = underlyingFA(source); err != nil {
		return fmt.Errorf("failed to parse input: %s\n", err.Error())
	}

	pBals, err := queryBalances(source)
	if err != nil {
		return fmt.Errorf("failed to get asset balance: %s", err.Error())
	}

	if pBals[tx.Input.Type] < tx.Input.Amount {
		return fmt.Errorf("not enough %s to cover the transaction", tx.Input.Type.String())
	}

	return nil
}

func ticker(asset string) (fat2.PTicker, error) {
	// No asset starts with a 'p', so we can do the quick check
	// if the start is a p for if it is already in 'p' form.
	// TODO: Make a more robust Asset -> pAsset converter
	if strings.ToUpper(asset) != "PEG" && asset[0] != 'p' {
		asset = "p" + strings.ToUpper(asset)
	}
	aType := fat2.StringToTicker(asset)
	if aType == fat2.PTickerInvalid {
		return fat2.PTickerInvalid, fmt.Errorf("invalid ticker type\n")
	}
	return aType, nil
}

// printFeWarning will tell the user of the Fe->Fa thing. Optionally provide a
// custom message format
func printFeWarning(cmd *cobra.Command, addrs ...string) bool {
	if q, _ := cmd.Flags().GetBool("no-warn"); q {
		return false
	}

	var warnings []string

	for _, addr := range addrs {
		add, addType, _ := underlyingFAWithType(addr)
		if addType == ADD_Fe || addType == ADD_FE {
			warnings = append(warnings, fmt.Sprintf("%s may be displayed as \n%s in some contexts such as explorers and cli outputs.", addr, add.String()))
		}
	}

	if len(warnings) > 0 {
		fmt.Print("\nNOTICE:\n\n")
		fmt.Print("DO NOT USE THE FOLLOWING FA ADDRESSES DIRECTLY, AS LOSS OF FUNDS MAY RESULT!\n\n")
		warns := strings.Join(warnings, "\n\n")
		fmt.Println(warns)
		fmt.Print("\nDO NOT USE THE FOLLOWING FA ADDRESSES DIRECTLY, AS LOSS OF FUNDS MAY RESULT!\n\n")
		fmt.Println("For more information about Ethereum backed addresses see:\n" +
			"https://example.com/link-to-info.html")
	}

	return false
}

func underlyingFA(addr string) (factom.FAAddress, error) {
	add, _, err := underlyingFAWithType(addr)
	return add, err
}

func underlyingFAWithType(addr string) (factom.FAAddress, int, error) {
	if len(addr) < 2 {
		add, err := factom.NewFAAddress(addr)
		return add, -1, err
	}
	switch addr[:2] {
	case "FA":
		// Resort to the default
	case "Fe":
		feAddr, err := factom.NewFeAddress(addr)
		return factom.FAAddress(feAddr), ADD_Fe, err
	case "FE":
		gatewayAddr, err := factom.NewFEGatewayAddress(addr)
		return factom.FAAddress(gatewayAddr), ADD_FE, err
	}
	add, err := factom.NewFAAddress(addr)
	return add, -1, err
}
