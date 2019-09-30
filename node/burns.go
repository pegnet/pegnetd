package node

import "encoding/hex"

var (
	// Pegnet Burn Addresses
	BurnAddresses = map[string]string{
		"MainNet": "EC2BURNFCT2PEGNETooo1oooo1oooo1oooo1oooo1oooo19wthin",
		//TestNetwork:    "EC2BURNFCT2TESTxoooo1oooo1oooo1oooo1oooo1oooo1EoyM6d",
	}

	BurnRCDs = map[string][32]byte{}
)

func init() {
	mr, _ := hex.DecodeString("37399721298d77984585040ea61055377039a4c3f3e2cd48c46ff643d50fd64f")
	var rcd [32]byte
	copy(rcd[:], mr[:])
	BurnRCDs["MainNet"] = rcd
}

func PegnetBurnAddress(network string) string {
	return BurnAddresses[network]
}

func PegnetBurnRCD(network string) [32]byte {
	return BurnRCDs[network]
}
