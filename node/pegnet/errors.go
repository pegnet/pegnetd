package pegnet

import "errors"

var (
	InsufficientBalanceErr          = errors.New("insufficient balance")
	InsufficientBalanceErrInt int64 = -1

	// -2 is an invalid tx. Usually by timestamps

	PFCTOneWayError          = errors.New("pFCT conversions are one way only at this height, they cannot be a conversion destination")
	PFCTOneWayErrorInt int64 = -3
	ZeroRatesError           = errors.New("an asset in the conversion has a rate of 0, and not allowed to be used for conversions")
	ZeroRatesErrorInt  int64 = -4
)

// IsRejectedTx takes an error, and returns the integer form of that error
// if it is a rejected tx. If the error is unknown, the original error is
// returned.
func IsRejectedTx(err error) (int64, error) {
	if err == nil {
		return 1, nil // No error!
	}
	if err == InsufficientBalanceErr {
		return InsufficientBalanceErrInt, nil
	}
	if err == PFCTOneWayError {
		return PFCTOneWayErrorInt, nil
	}
	if err == ZeroRatesError {
		return ZeroRatesErrorInt, nil
	}
	return 0, err
}
