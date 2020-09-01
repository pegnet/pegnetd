package node

import (
	"encoding/hex"

	"github.com/Factom-Asset-Tokens/factom"
)

var (
	// BurnAddress is the mainnet burn address
	BurnAddress = "EC2BURNFCT2PEGNETooo1oooo1oooo1oooo1oooo1oooo19wthin"

	// BurnAddress that can be used for all assets
	GlobalBurnAddress = "FA2BURNBABYBURNoooooooooooooooooooooooooooooooDGvNXy"

	// Global Burn Address as proper FAAddress
	FAGlobalBurnAddress factom.FAAddress

	// BurnRCD is the rcd representation of the burn address
	BurnRCD = [32]byte{}
)

func init() {
	mr, _ := hex.DecodeString("37399721298d77984585040ea61055377039a4c3f3e2cd48c46ff643d50fd64f")
	copy(BurnRCD[:], mr[:])
}
