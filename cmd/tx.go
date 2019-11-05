package cmd

import (
	"fmt"
	"strings"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"
	"github.com/pegnet/pegnetd/node"
)

func signAndSend(tx *fat2.Transaction, cl *factom.Client, payment string) (err error, commit *factom.Bytes32, reveal *factom.Bytes32) {
	// Get out private key
	priv, err := tx.Input.Address.GetFsAddress(cl)
	if err != nil {
		return fmt.Errorf("unable to get private key: %s\n", err.Error()), nil, nil
	}

	var txBatch fat2.TransactionBatch
	txBatch.Version = 1
	txBatch.Transactions = []fat2.Transaction{*tx}
	txBatch.ChainID = &node.TransactionChain

	// Sign the tx and make an entry
	err = txBatch.MarshalEntry()
	if err != nil {
		return fmt.Errorf("failed to marshal tx: %s", err.Error()), nil, nil
	}

	txBatch.Sign(priv)

	if err := txBatch.Validate(); err != nil {
		return fmt.Errorf("invalid tx: %s", err.Error()), nil, nil
	}

	ec, err := factom.NewECAddress(payment)
	if err != nil {
		return fmt.Errorf("failed to parse input: %s\n", err.Error()), nil, nil
	}

	bal, err := ec.GetBalance(cl)
	if err != nil {
		return fmt.Errorf("failed to get ec balance: %s\n", err.Error()), nil, nil
	}

	if cost, err := txBatch.Cost(false); err != nil || uint64(cost) > bal {
		return fmt.Errorf("not enough ec balance for the transaction"), nil, nil
	}

	es, err := ec.GetEsAddress(cl)
	if err != nil {
		return fmt.Errorf("failed to parse input: %s\n", err.Error()), nil, nil
	}

	txid, err := txBatch.ComposeCreate(cl, es, false)
	if err != nil {
		return fmt.Errorf("failed to submit entry: %s\n", err.Error()), nil, nil
	}

	return nil, txid, txBatch.Hash
}

func setTransferOutput(tx *fat2.Transaction, cl *factom.Client, dest, amt string) error {
	var err error
	amount, err := FactoidToFactoshi(amt)
	if err != nil {
		return fmt.Errorf("invalid amount specified: %s\n", err.Error())
	}

	tx.Transfers = make([]fat2.AddressAmountTuple, 1)
	tx.Transfers[0].Amount = uint64(amount)
	if tx.Transfers[0].Address, err = factom.NewFAAddress(dest); err != nil {
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
	if tx.Input.Address, err = factom.NewFAAddress(source); err != nil {
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
