package fat2_test

import (
	"encoding/json"
	"testing"

	"github.com/Factom-Asset-Tokens/factom"

	. "github.com/pegnet/pegnetd/fat/fat2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Factoid key pairs used in this set of tests:
// Sands: Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK -- FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q
// Zeros: Fs1KWJrpLdfucvmYwN2nWrwepLn8ercpMbzXshd1g8zyhKXLVLWj -- FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC

var validTransactionBatchJSON = `{
	"version": 1,
	"transactions": [{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
	}]
}`

var transactionBatchUnmarshalJSONTests = []struct {
	Name   string
	Error  string
	TxJSON string
}{{
	Name:   "valid batch",
	TxJSON: validTransactionBatchJSON,
}, {
	Name:  "double version",
	Error: "*fat2.TransactionBatch: unexpected JSON length",
	TxJSON: `{
		"version": 1,
		"version": 1,
		"transactions": [{
			"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
			"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
		}]
	}`,
}, {
	Name:  "double transactions",
	Error: "*fat2.TransactionBatch: unexpected JSON length",
	TxJSON: `{
		"version": 1,
		"transactions": [{
			"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
			"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
		}],
		"transactions": [{
			"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
			"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
		}]
	}`,
}, {
	Name:  "extra random field",
	Error: "*fat2.TransactionBatch: unexpected JSON length",
	TxJSON: `{
		"version": 1,
		"transactions": [{
			"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
			"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
		}],
		"BAD": "BAD"
	}`,
}}

// TestTransactionBatch_UnmarshalJSON tests the UnmarshalJSON() function, which
// checks that a given JSON byte array is a valid TransactionBatch and contains
// no duplicate or extra JSON keys
func TestTransactionBatch_UnmarshalJSON(t *testing.T) {
	for _, test := range transactionBatchUnmarshalJSONTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			var txBatch TransactionBatch
			err := json.Unmarshal([]byte(test.TxJSON), &txBatch)
			if len(test.Error) != 0 {
				assert.EqualError(err, test.Error)
				return
			}
			assert.Nil(err)
		})
	}
}

var transactionBatchValidateTests = []struct {
	Name   string
	Error  string
	Keys   []string
	TxJSON string
}{{
	Name:   "valid batch (all required signatures)",
	Keys:   []string{"Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK"},
	TxJSON: validTransactionBatchJSON,
}, {
	Name:   "missing signature",
	Error:  "invalid number of ExtIDs",
	Keys:   []string{},
	TxJSON: validTransactionBatchJSON,
}, {
	Name:  "repeat signature",
	Error: "invalid number of ExtIDs",
	Keys: []string{
		"Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK",
		"Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK",
	},
	TxJSON: validTransactionBatchJSON,
}, {
	Name:  "extra signature",
	Error: "invalid number of ExtIDs",
	Keys: []string{
		"Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK",
		"Fs1KWJrpLdfucvmYwN2nWrwepLn8ercpMbzXshd1g8zyhKXLVLWj",
	},
	TxJSON: validTransactionBatchJSON,
}, {
	Name:   "extra signature",
	Error:  "ExtIDs[1]: unexpected or duplicate RCD Hash",
	Keys:   []string{"Fs1KWJrpLdfucvmYwN2nWrwepLn8ercpMbzXshd1g8zyhKXLVLWj"},
	TxJSON: validTransactionBatchJSON,
}}

// TestTransactionBatch_Validate tests the Validate() function, which performs all validation checks
// on a given TransactionBatch struct. However the main purpose is to test that the signature creation
// and validation performs as expected. The actual transaction validation is tested more thoroughly in
// the tests for the Transaction struct.
func TestTransactionBatch_Validate(t *testing.T) {
	for _, test := range transactionBatchValidateTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			var txBatch TransactionBatch
			err := json.Unmarshal([]byte(test.TxJSON), &txBatch)
			require.NoError(t, err, "TransactionBatch JSON Unmarshal raised an unexpected error")

			c := factom.NewBytes32("00000000000000000000000000000000")
			txBatch.Entry.ChainID = &c
			signerKeys := make([]factom.RCDSigner, len(test.Keys))
			for i, keyString := range test.Keys {
				key := factom.FsAddress{}
				err = key.Set(keyString)
				signerKeys[i] = key
				require.NoError(t, err, "FsAddress.Set()")
			}
			ent, err := txBatch.Sign(signerKeys...)
			assert.NoError(err)
			txBatch.Entry = ent

			err = txBatch.Validate()
			if len(test.Error) != 0 {
				assert.EqualError(err, test.Error)
				return
			}
			assert.Nil(err)
		})
	}
}
