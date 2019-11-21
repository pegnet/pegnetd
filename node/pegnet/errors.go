package pegnet

import "errors"

var (
	InsufficientBalanceErr = errors.New("insufficient balance")
	PFCTOneWayError        = errors.New("pFCT conversions are one way only at this height, they cannot be a conversion destination")
)
