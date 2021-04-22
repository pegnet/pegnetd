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

	ratesOverPeriod := d.LastAveragesData //                    First collect all the values over the blocks
	averages := map[fat2.PTicker]uint64{} //                      in the average period, then compute the averages
	if ratesOverPeriod == nil {           //                    If no map exists yet
		ratesOverPeriod = map[fat2.PTicker][]uint64{} //          create one.
	}

	defer func() { //                                           Always set up the cache when exiting the routine
		d.LastAveragesData = ratesOverPeriod //                   Save the data we used to create averages
		d.LastAveragesHeight = height        //                   Save the height of this data
		d.LastAverages = averages            //                   Save the averages we computed
	}()

	// collectRatesAtHeight
	// This routine collects all the data used to compute an average.  If any data is missing, then
	// that data is represented by a zero.
	collectRatesAtHeight := func(h uint32) {
		for k := range ratesOverPeriod { //                   Make sure there is room for a new height
			for len(ratesOverPeriod[k]) >= int(AveragePeriod) { //  If at the limit or above,
				copy(ratesOverPeriod[k], ratesOverPeriod[k][1:])                    // Shift data down 1 element
				ratesOverPeriod[k] = ratesOverPeriod[k][:len(ratesOverPeriod[k])-1] //   And drop off the last value
			}
		}

		if rates, err := d.Pegnet.SelectRates(ctx, h); err != nil { // Pull the rates out of the database at each
			panic("no recovery from a database error getting rates")
		} else {
			for k, v := range rates { //                            For all the rates
				if ratesOverPeriod[k] == nil { //                     if no rates yet, at a slice for them
					ratesOverPeriod[k] = []uint64{} //              Allocate the slice
				}
				ratesOverPeriod[k] = append(ratesOverPeriod[k], v) // Add the rates we find
			}
		}
	}

	switch {
	//   If the LastAveragesHeight is out of range given our current height, we just need to load
	//     all the values
	case d.LastAveragesHeight+1 < height || d.LastAveragesHeight > height:
		for k, v := range ratesOverPeriod {
			if v != nil {
				ratesOverPeriod[k] = ratesOverPeriod[k][:0]
			}
		}
		startHeightS := int64(height) - (int64(AveragePeriod)) + 1 // startHeight is AveragePeriod before height+1
		//                                                            (add 1 so the block at height is included)
		if startHeightS < 1 { //                                    If AveragePeriod blocks don't exist,
			startHeightS = 1 //                                      then flour the start to 1
		}

		startHeight := uint32(startHeightS)

		for h := startHeight; h <= height; h++ { //            Collect rates over the blocks (including height)
			collectRatesAtHeight(h) //                           and add them to ratesOverPeriod
		}

	//   If all we need is the next height, then only collect that height.
	case d.LastAveragesHeight+1 == height:
		collectRatesAtHeight(height) //                         Add the current height to the dataset so far
	}

	for k, v := range ratesOverPeriod { //                        The average rate is zero for any asset without
		averages[k] = 0                                       //    the number of required rates
		if AveragePeriod-numberMissing(v) < AverageRequired { //  Count the missing values, and if not enough
			continue //                                             skip it
		}
		for _, v2 := range v { //                               Sum up all the rates found for an asset
			averages[k] += v2 //                                The assumption is that rates are no where near
		} //                                                       64 bits, so they won't overflow
		averages[k] = averages[k] / uint64(len(v)) // Divide the sum of the rates by the number of rates
	}

	return averages // Return the rates we found.
}

func numberMissing(dataset []uint64) (numZeros uint64) {
	for _, v := range dataset {
		if v == 0 {
			numZeros++
		}
	}
	if len(dataset) < int(AveragePeriod) {
		numZeros += AveragePeriod - uint64(len(dataset))
	}
	return numZeros
}
