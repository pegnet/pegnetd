package srv

import "github.com/Factom-Asset-Tokens/factom"

// underlyingFA will return the FA address given an input of type
// FA, Fe, or FE
func underlyingFA(addr string) (factom.FAAddress, error) {
	if len(addr) < 2 {
		add, err := factom.NewFAAddress(addr)
		return add, err
	}
	switch addr[:2] {
	case "FA":
		// Resort to the default
	case "Fe":
		feAddr, err := factom.NewFeAddress(addr)
		return factom.FAAddress(feAddr), err
	case "FE":
		gatewayAddr, err := factom.NewFEGatewayAddress(addr)
		return factom.FAAddress(gatewayAddr), err
	}
	add, err := factom.NewFAAddress(addr)
	return add, err
}
