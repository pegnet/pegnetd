package pegnet

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

const (
	DefaultPad = 3
)

// SplitTxID splits a TxID into it's parts.
// TxID format : [TxIndex]-[EntryHash]
//				 1-c99dedea0e4e0c40118fe7e4d515b23cc0489269c8cef187b4f15a4ccbd880be
func SplitTxID(txid string) (index int, entryhash string, err error) {
	arr := strings.Split(txid, "-")
	if len(arr) != 2 {
		return -1, "", fmt.Errorf("txid does not match txid format, format: [TxIndex]-[EntryHash]")
	}

	txIndex, err := strconv.ParseInt(arr[0], 10, 32)
	if err != nil {
		return -1, "", fmt.Errorf("index must be a valid integer")
	}
	if txIndex < 0 {
		return -1, "", fmt.Errorf("index must be a postive integer")
	}

	// Verify the entryhash is valid
	hash, err := hex.DecodeString(arr[1])
	if err != nil {
		return -1, "", fmt.Errorf("entryhash must be a valid hex string")
	}
	if len(hash) != 32 {
		return -1, "", fmt.Errorf("entryhash must be 32 bytes (64 hex characters)")
	}

	// Over 10k index? That's impossible, that means each tx is of 1 byte
	if txIndex > 10000 {
		return -1, "", fmt.Errorf("txindex is too large to be valid")
	}
	return int(txIndex), arr[1], nil
}

// FormatTxID constructs a txid from an entryhash and it's index
func FormatTxID(index int, hash string) string {
	return FormatTxIDWithPad(DefaultPad, index, hash)
}

// FormatTxIDWithPad constructs a txid from an entryhash and it's index.
// It will pad the index such that it is of at least 'pad' characters in lenght.
// pad = 2 -> 01-entryhash
// pad = 3 -> 001-entryhash
func FormatTxIDWithPad(pad, index int, hash string) string {
	// format is the "%0Nd-%s
	format := fmt.Sprintf("%%0%dd-%%s", pad)
	return fmt.Sprintf(format, index, hash)
}
