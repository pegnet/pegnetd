package pegnet_test

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/pegnet/pegnet/modules/conversions"
	. "github.com/pegnet/pegnetd/node/pegnet"
	"github.com/stretchr/testify/assert"
)

func TestFormatTxID(t *testing.T) {
	assert := assert.New(t)

	// Ensure formats are valid
	t.Run("default", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			data := make([]byte, 32)
			rand.Read(data)
			idx, hash := rand.Intn(10000), hex.EncodeToString(data)
			txid := FormatTxID(idx, hash)

			fIdx, fHash, err := SplitTxID(txid)
			assert.NoError(err)
			assert.Equal(idx, fIdx)
			assert.Equal(hash, fHash)
		}
	})

	// Ensure formats are valid with different pads
	t.Run("random pad", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			data := make([]byte, 32)
			rand.Read(data)
			idx, hash := rand.Intn(10000), hex.EncodeToString(data)
			txid := FormatTxIDWithPad(rand.Intn(40), idx, hash)

			fIdx, fHash, err := SplitTxID(txid)
			assert.NoError(err)
			assert.Equal(idx, fIdx)
			assert.Equal(hash, fHash)
		}
	})
}

func TestVerifyTransactionHash(t *testing.T) {
	assert := assert.New(t)
	type TestVec struct {
		TxID string
		// If Error is set, an error is expected
		Error     string
		EntryHash string // If this is set, the entryhash and txindex are checked
		TxIndex   int
		Pad       int
	}
	vects := []TestVec{
		{TxID: "0-", Error: "tx has no entryhash"},
		{TxID: "0-aa", Error: "entryhash too short"},
		{TxID: "0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a", Error: "entryhash too short"},
		{TxID: "0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25aaaaa", Error: "entryhash too long"},
		{TxID: "0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a5", Error: "entryhash odd character length"},
		{TxID: "-2-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", Error: "negative, and too many splits"},
		{TxID: "a2-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", Error: "txindex not a number"},
		{TxID: "179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", Error: "hash too short"},
		{TxID: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57aa", Error: "hash too long"},

		// Valids
		{TxID: "0-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 0, Pad: 1},
		{TxID: "0010-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 10, Pad: 4},
		{TxID: "012-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 12, Pad: 3},
		{TxID: "9-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 9, Pad: 1},
		{TxID: "00000-17c05acb2fec5add1bfadc4c5d7fbd532a1a3fdad0b7b8dee97a544b4ab77396",
			EntryHash: "17c05acb2fec5add1bfadc4c5d7fbd532a1a3fdad0b7b8dee97a544b4ab77396", TxIndex: 0, Pad: 5},
		{TxID: "999999-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57"},

		// Test some under valued pads
		{TxID: "999999-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 999999, Pad: 0},
		{TxID: "12345-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 12345, Pad: 1},
		{TxID: "12-1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: 12, Pad: 1},

		// Test Batch Hashes
		{TxID: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57",
			EntryHash: "1a179409cc789a3eb1061e6b7c783c622c39d5bd78e6fd0aca2a13c0e1a25a57", TxIndex: -1},
	}

	for i := range vects {
		vec := vects[i]
		index, entryhash, err := VerifyTransactionHash(vec.TxID)
		if err != nil && vec.Error == "" {
			// Error not expected!
			t.Errorf("%d: should be good, found err: %s", i, err.Error())
		} else if err == nil && vec.Error != "" {
			t.Errorf("%d: found no error, expected one. %s", i, vec.Error)
		} else if err == nil && vec.Error == "" {
			if vec.EntryHash != "" && index != vec.TxIndex {
				t.Errorf("exp index of %d, found %d", vec.TxIndex, index)
			}
			if vec.EntryHash != "" && vec.EntryHash != entryhash {
				t.Errorf("exp ehash of %s, found %s", vec.EntryHash, entryhash)
			}

			// -1 tx idx means don't check the reformatting
			if vec.EntryHash != "" && vec.TxIndex != -1 {
				exp := FormatTxIDWithPad(vec.Pad, vec.TxIndex, vec.EntryHash)
				assert.Equal(exp, vec.TxID)
			}
		} else {
		}
	}
}

func TestPEGSupplyConversions(t *testing.T) {
	min := func(a, b int64) int64 {
		if a < b {
			return a
		}
		return b
	}

	// Currently the PEG supply limit yields are calculated as such:
	// amt pXXX -> yielded PEG + refund pXXX
	t.Run("test equivalency", func(t *testing.T) {
		for i := 0; i < 1; i++ {
			amtR := rand.Uint64() % (5 * 1e6 * 1e8) // 50K max
			pegR := rand.Uint64() % (5 * 1e6 * 1e8) // 50K max

			input := rand.Int63() % (1 * 1e6 * 1e8) // 1million max
			_, err := conversions.Convert(int64(input), amtR, pegR)
			if err != nil {
				continue // Likely an overflow or rate is 0
			}

			// Most yield possibilities for a 5K bank
			for yield := int64(1); yield <= min(input, 5000*1e8); yield = yield + (rand.Int63() % 1e8) {
				// 2 methods to calculate the refund. We have:
				// Input in pXXX, yield in PEG

				refund := RefundMethod2(t, input, yield, amtR, pegR)
				CheckRefund(t, input, refund, yield, amtR, pegR)
			}
		}
	})
}

// RefundMethod1 is the following:
// maxPEGYield := (input -> PEG)
// refundPEG := maxPEGYield - PEGYield
// refuind := (refundPEG -> pXXX)
func RefundMethod1(t *testing.T, input, pegYield int64, amtRate, pegRate uint64) int64 {
	maxPEGYield, _ := conversions.Convert(input, amtRate, pegRate)
	refundPEG := maxPEGYield - pegYield
	refund, _ := conversions.Convert(refundPEG, pegRate, amtRate)
	return refund
}

// RefundMethod2 is the following:
// consumedInput := (pegYield -> pXXX)
// refund := input - consumedInput
func RefundMethod2(t *testing.T, input, pegYield int64, amtRate, pegRate uint64) int64 {
	consumedInput, _ := conversions.Convert(pegYield, pegRate, amtRate)
	refund := input - consumedInput
	return refund
}

// CheckRefund
// amt is in pXXX
// refund is in pXXX
// pegYield is in PEG
func CheckRefund(t *testing.T, input, refund, pegYield int64, amtRate, pegRate uint64) {
	maxPegYield, err := conversions.Convert(input, amtRate, pegRate)
	if err != nil {
		return // Overflow or 0 rates
	}

	{
		// Asset Equivalency
		// This check is `input = refund + (peg converted to input)`
		yieldInAsset, err := conversions.Convert(pegYield, pegRate, amtRate)
		if err != nil {
			t.Error(err) // This would be bad news
		}

		if refund+yieldInAsset != int64(input) {
			t.Errorf("input = refund + (yield PEG -> pXXX) does not hold true\n"+
				"Amt: %d, Refund: %d, Add: %d\n"+
				"Difference: %d", input, refund, yieldInAsset, int64(input)-(refund+yieldInAsset))
		}
	}

	{
		// PEG Equivalency
		// This check is
		// consumed = input - refund
		// consumed -> PEG + refund -> PEG = input -> PEG
		consumed := int64(input) - refund
		consumedPEG, err := conversions.Convert(consumed, amtRate, pegRate)
		if err != nil {
			t.Error(err) // This would be bad news
		}

		refundPEGCheck, err := conversions.Convert(refund, amtRate, pegRate)
		if err != nil {
			t.Error(err) // This would be bad news
		}

		// We allow a difference of +1. This means the consumed + refund is
		// 1 less than the max. Which is ok, and expected
		if maxPegYield-(consumedPEG+refundPEGCheck) > 1 {
			t.Errorf("Failed PEG equivalency: %d", maxPegYield-(consumedPEG+refundPEGCheck))
		}
	}
}
