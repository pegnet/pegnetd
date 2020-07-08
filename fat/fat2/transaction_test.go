package fat2_test

import (
	"encoding/json"
	"testing"

	. "github.com/pegnet/pegnetd/fat/fat2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var addressAmountTupleUnmarshalJSONTests = []struct {
	Name  string
	Error string
	JSON  string
}{{
	Name: "valid",
	JSON: `{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}`,
}, {
	Name:  "double address",
	Error: "*fat2.AddressAmountTuple: unexpected JSON length",
	JSON:  `{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}`,
}, {
	Name:  "double amount",
	Error: "*fat2.AddressAmountTuple: unexpected JSON length",
	JSON:  `{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50, "amount": 50}`,
}, {
	Name:  "extra random field",
	Error: "*fat2.AddressAmountTuple: unexpected JSON length",
	JSON:  `{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50, "BAD": "BAD"}`,
}, {
	Name:  "empty",
	Error: "*fat2.AddressAmountTuple.Address: unexpected end of JSON input",
	JSON:  `{}`,
}}

// TestAddressAmountTuple_UnmarshalJSON tests the UnmarshalJSON() function,
// which checks that a given JSON byte array is a valid AddressAmountTuple
// and contains no duplicate or extra JSON keys
func TestAddressAmountTuple_UnmarshalJSON(t *testing.T) {
	for _, test := range addressAmountTupleUnmarshalJSONTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			var tuple AddressAmountTuple
			err := json.Unmarshal([]byte(test.JSON), &tuple)
			if len(test.Error) != 0 {
				assert.EqualError(err, test.Error)
				return
			}
			assert.Nil(err)
		})
	}
}

var typedAddressAmountTupleUnmarshalJSONTests = []struct {
	Name  string
	Error string
	JSON  string
}{{
	Name: "valid",
	JSON: `{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50, "type": "PEG"}`,
}, {
	Name:  "double address",
	Error: "*fat2.TypedAddressAmountTuple: unexpected JSON length",
	JSON:  `{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50, "type": "PEG"}`,
}, {
	Name:  "double amount",
	Error: "*fat2.TypedAddressAmountTuple: unexpected JSON length",
	JSON:  `{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50, "amount": 50, "type": "PEG"}`,
}, {
	Name:  "double type",
	Error: "*fat2.TypedAddressAmountTuple: unexpected JSON length",
	JSON:  `{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50, "type": "pUSD", "type": "PEG"}`,
}, {
	Name:  "extra random field",
	Error: "*fat2.TypedAddressAmountTuple: unexpected JSON length",
	JSON:  `{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50, "type": "PEG", "BAD": "BAD"}`,
}, {
	Name:  "empty",
	Error: "*fat2.TypedAddressAmountTuple.Address: unexpected end of JSON input",
	JSON:  `{}`,
}}

// TestTypedAddressAmountTuple_UnmarshalJSON tests the UnmarshalJSON() function,
// which checks that a given JSON byte array is a valid TypedAddressAmountTuple
// and contains no duplicate or extra JSON keys
func TestTypedAddressAmountTuple_UnmarshalJSON(t *testing.T) {
	for _, test := range typedAddressAmountTupleUnmarshalJSONTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			var tuple TypedAddressAmountTuple
			err := json.Unmarshal([]byte(test.JSON), &tuple)
			if len(test.Error) != 0 {
				assert.EqualError(err, test.Error)
				return
			}
			assert.Nil(err)
		})
	}
}

var transactionUnmarshalJSONTests = []struct {
	Name   string
	Error  string
	TxJSON string
}{{
	Name: "valid transfer",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
	}`,
}, {
	Name: "valid transfer with metadata",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}],
		"metadata": {"foo": "bar", "baz": 123}
	}`,
}, {
	Name: "valid conversion",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"conversion": "pUSD"
	}`,
}, {
	Name:   "empty",
	Error:  "*fat2.Transaction.Input: unexpected end of JSON input",
	TxJSON: `{}`,
}, {
	Name:  "double input",
	Error: "*fat2.Transaction: unexpected JSON length",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
	}`,
}, {
	Name:  "double transfers",
	Error: "*fat2.Transaction: unexpected JSON length",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}],
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
	}`,
}, {
	Name:  "double conversion",
	Error: "*fat2.Transaction: unexpected JSON length",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"conversion": "pUSD",
		"conversion": "PEG"
	}`,
}, {
	Name:  "combined transfers and conversion",
	Error: "*fat2.Transaction.Input: *fat2.TypedAddressAmountTuple.Amount: unexpected end of JSON input",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG"},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}],
		"conversion": "pUSD"
	}`,
}, {
	Name:  "double metadata",
	Error: "*fat2.Transaction: unexpected JSON length",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"conversion": "pUSD",
		"metadata": {"foo": "bar", "baz": 123},
		"metadata": {"foo": "bar", "baz": 123}
	}`,
}, {
	Name:  "extra random field",
	Error: "*fat2.Transaction: unexpected JSON length",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"conversion": "pUSD",
		"BAD": "BAD"
	}`,
}}

// TestTransaction_UnmarshalJSON tests the UnmarshalJSON() function, which
// checks that a given JSON byte array is a valid Transaction and contains
// no duplicate or extra JSON keys
func TestTransaction_UnmarshalJSON(t *testing.T) {
	for _, test := range transactionUnmarshalJSONTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			var tx Transaction
			err := json.Unmarshal([]byte(test.TxJSON), &tx)
			if len(test.Error) != 0 {
				assert.EqualError(err, test.Error)
				return
			}
			assert.Nil(err)
		})
	}
}

var transactionValidationTests = []struct {
	Name   string
	Error  string
	TxJSON string
}{{
	Name: "valid transfer",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
	}`,
}, {
	Name: "valid transfers",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 100},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}, {"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
	}`,
}, {
	Name: "valid conversion",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"conversion": "pUSD"
	}`,
}, {
	Name:  "illegal use of burn address",
	Error: "invalid input: FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC reserved as burn address",
	TxJSON: `{
		"input": {"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "type": "PEG", "amount": 50},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
	}`,
}, {
	Name:  "empty transfers",
	Error: "at least one transfer or exactly one conversion type required",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"transfers": []
	}`,
}, {
	Name:  "insufficient input",
	Error: "insufficient input to cover outputs",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 99},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}, {"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 50}]
	}`,
}, {
	Name:  "input != sum of transfers",
	Error: "input amount must equal sum of transfer amounts",
	TxJSON: `{
		"input": {"address": "FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q", "type": "PEG", "amount": 50},
		"transfers": [{"address": "FA1zT4aFpEvcnPqPCigB3fvGu4Q4mTXY22iiuV69DqE1pNhdF2MC", "amount": 25}]
	}`,
}}

// TestTransaction_Validate tests that the Validate function on a Transaction struct is able to catch all
// formatting errors on a given transaction's input and outputs.
func TestTransaction_Validate(t *testing.T) {
	for _, test := range transactionValidationTests {
		t.Run(test.Name, func(t *testing.T) {
			assert := assert.New(t)
			var tx Transaction
			err := json.Unmarshal([]byte(test.TxJSON), &tx)
			require.NoErrorf(t, err, "Transaction JSON Unmarshal raised an unexpected error: %v", err)
			err = tx.Validate()
			if len(test.Error) != 0 {
				assert.EqualError(err, test.Error)
				return
			}
			assert.Nil(err)
		})
	}
}
