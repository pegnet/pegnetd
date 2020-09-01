package node

import (
	"bytes"
	"crypto/sha256"
)

//ComputeChainIDFromFields
// Takes the binary fields that define a chainID and hashes
// them together to create the ChainID expected by the APIs.
// These fields are treated as binary, but could be simple text, like:
// "Bob's" "Favorite" "Chain"
func ComputeChainIDFromFields(fields [][]byte) [32]byte {
	var buf bytes.Buffer
	for _, id := range fields {
		h := sha256.Sum256(id)
		_, err := buf.Write(h[:])
		if err != nil {
			panic("Unexpected write failure")
		}
	}
	return sha256.Sum256(buf.Bytes())
}

// ComputeChainIDFromStrings
// Take a set of strings, and compute the chainID.  If you have binary fields, you
// can call factom.ComputeChainIDFromFields directly.
func ComputeChainIDFromStrings(fields []string) [32]byte {
	var binary [][]byte
	for _, str := range fields {
		binary = append(binary, []byte(str))
	}
	return ComputeChainIDFromFields(binary)
}
