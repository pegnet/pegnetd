package conversions

import (
	"fmt"
	"math/big"

	"github.com/pegnet/pegnetd/config"
)

// Convert
// takes an input amount and returns an output amount that can be created
// from it, given the two rates `fromRate` and `toRate` denominated in 1e-8 USD.
// All parameters must be in their lowest divisible unit as whole numbers.
//
// Prior to PIP-10:
//  RateSource = fromRate
//  RateDest = toRate
//
// Under PIP-10:
//  RateSource = min(fromRate, fromAvg)
//  RateDest = max(toRate, toAvg)
//
//  A Source Tokens
//  X Destination Tokens
//
//     A           fromRate USD         1
//  ----------  *  ------------  *  ----------  =  X
//     1            1 fromType      toRate USD
//
func Convert(height uint32, amount int64, fromRate, fromAvg, toRate, toAvg uint64) (int64, error) {
	if amount < 0 { //                                           Must have something to convert; negative numbers
		return 0, fmt.Errorf("invalid amount: must be greater than or equal to zero") // do not qualify
	}

	if fromRate == 0 || toRate == 0 { //                         If either rate (fromRate or toRate) is zero
		return 0, fmt.Errorf("invalid rate: 0") //                  then we can't do the conversion
	}
	RateSource := fromRate
	RateDest := toRate
	if height >= config.PIP10AverageActivation { //              Check that PIP-10 has been activated.
		if fromAvg == 0 || toAvg == 0 { //                       If either Average (fromAvg or toAvg) is zero
			return 0, fmt.Errorf("invalid rate: 0") //              then we can't do the conversion
		}
		if fromRate > fromAvg { //                               Use the min(fromRate, fromAvg)
			RateSource = fromAvg
		}
		if toRate < toAvg { //                                   Use the max(toRate,toAvg)
			RateDest = toAvg
		}
	}

	// Convert the rates to integers. Because these rates are in USD, we will switch all our inputs to
	// 1e-8 fixed point. The `want` should already be in this format. This should be the most amount of
	// accuracy a miner reports. Anything beyond the 8th decimal point, we cannot account for.
	//
	// Uses big ints to avoid overflows.
	fr := new(big.Int).SetUint64(RateSource)
	tr := new(big.Int).SetUint64(RateDest)
	amt := big.NewInt(amount)

	// Now we can run the conversion
	// ALWAYS multiply first. If you do not adhere to the order of operations shown
	// explicitly below, your answer will be incorrect. When doing a conversion,
	// always multiply before you divide.
	//  (amt * fromrate) / torate
	num := big.NewInt(0).Mul(amt, fr)
	num.Div(num, tr)
	if !num.IsInt64() {
		return 0, fmt.Errorf("integer overflow")
	}
	return num.Int64(), nil
}
