package node

import (
	"context"

	"github.com/pegnet/pegnetd/fat/fat2"
)

const AveragePeriod = uint64(288)         // Our Average Period is 2 days (144 10 minute blocks per day)
const AverageRequired = AveragePeriod / 2 // If we have at least half the rates, we can do conversions

// getPegNetRateAverages
// Gets all the rates for the AveragePeriod (the number of blocks contributing to the average), and computes
// the average rate for all assets.  If values are missing for an asset for any of the blocks in the AveragePeriod,
// then don't allow conversions by setting the averate to zero for that asset
//
// Return a map of averages of the form map[fat2.PTicker]uint64
//
// Also note that if asked twice about the same height, we cache the response.
func (d *Pegnetd) GetPegNetRateAverages(ctx context.Context, height uint32) (Avg interface{}) {

	if d.LastAveragesHeight == height { //                      If a cache hit is detected, return the cache value
		return d.LastAverages
	}

	ratesOverPeriod := map[fat2.PTicker][]uint64{} //           First collect all the values over the blocks
	averages := map[fat2.PTicker]uint64{}          //             in the average period, then compute the averages

	defer func() { //                                           Always set up the cache when exiting the routine
		d.LastAveragesHeight = height
		d.LastAverages = averages
	}()

	startHeight := height - (uint32(AveragePeriod)) + 1 //      The startHeight is AveragePeriod before height+1
	//                                                            (add 1 so the block at height is included)
	if startHeight < 1 { //                                     If AveragePeriod blocks don't exist, then ignore
		return averages
	}

	for h := startHeight; height <= height; h++ { //            Collect rates over the blocks (including height)
		var rates map[fat2.PTicker]uint64 //                    Collect all the rates
		var err error

		if rates, err = d.Pegnet.SelectRates(ctx, h); err != nil { // Pull the rates out of the database at each
			return averages //                                          height.  Return averages if an error
		}

		for k, v := range rates { //                            For all the rates
			if ratesOverPeriod[k] == nil { //                     if no rates yet, at a slice for them
				ratesOverPeriod[k] = []uint64{} //              Allocate the slice
			}
			if v != 0 { //                                      Only collect non-zero rates
				ratesOverPeriod[k] = append(ratesOverPeriod[k], v) // Add the rates we find
			}
		}
	}

	for k, v := range ratesOverPeriod { //                      When we average the rates, we return a zero for
		averages[k] = 0                       //                  any asset that doesn't have a rate in all blocks
		if uint64(len(v)) < AverageRequired { //                We can see missing rates because the list isn't
			continue //                                           long enough. Too many missing rates, and we skip it
		}
		for _, v2 := range v { //                               Sum up all the rates found for an asset
			averages[k] += v2 //                                The assumption is that rates are no where near
		} //                                                       64 bits, so they won't overflow
		averages[k] = averages[k] / uint64(len(v)) // Divide the sum of the rates by the number of rates
	}

	return averages // Return the rates we found.
}
