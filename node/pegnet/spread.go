package pegnet

import "github.com/pegnet/pegnet/modules/conversions"

type Quote struct {
	MarketRate    uint64 `json:"marketrate"`
	MovingAverage uint64 `json:"movingaverage"`
}

func (q Quote) GetTolerance() uint64 {
	return q.MarketRate / SpreadToleranceFactor
}

func (q Quote) MinTolerance() uint64 {
	if q.MovingAverage >= q.MarketRate {
		return q.MarketRate
	}
	tolerance := q.GetTolerance()
	toleranced := q.MovingAverage + tolerance
	return min(q.MarketRate, toleranced)

}

func (q Quote) MaxTolerance() uint64 {
	if q.MovingAverage <= q.MarketRate {
		return q.MarketRate
	}

	tolerance := q.GetTolerance()
	toleranced := q.MovingAverage - tolerance
	if tolerance > q.MovingAverage {
		q.MovingAverage = 0 // Protect underflow
	}
	return max(q.MarketRate, toleranced)
}

func (q Quote) Spread() int64 {
	return spread(q.MarketRate, 1e8, q.Min(), 1e8)
}

func (q Quote) SpreadWithTolerance() int64 {
	return spread(q.MarketRate, 1e8, q.MinTolerance(), 1e8)
}

func (q Quote) MakeBase(quoteCurrency Quote) QuotePair {
	return QuotePair{
		BaseCurrency:  q,
		QuoteCurrency: quoteCurrency,
	}
}

func (q Quote) MakeQuote(baseCurrency Quote) QuotePair {
	return QuotePair{
		BaseCurrency:  baseCurrency,
		QuoteCurrency: q,
	}
}

func (q Quote) Max() uint64 {
	return max(q.MarketRate, q.MovingAverage)
}

func (q Quote) Min() uint64 {
	return min(q.MarketRate, q.MovingAverage)
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

type QuotePair struct {
	// Base/Quote
	// E.g BTC/USD
	BaseCurrency  Quote `json:"basecurrency"`
	QuoteCurrency Quote `json:"quotecurrency"`
}

func (q QuotePair) SellRate() int64 {
	rate, err := conversions.Convert(1e8, q.BaseCurrency.MinTolerance(), q.QuoteCurrency.MaxTolerance())
	if err != nil {
		return -1
	}
	return rate
}

func (q QuotePair) BuyRate() int64 {
	rate, err := conversions.Convert(1e8, q.BaseCurrency.MaxTolerance(), q.QuoteCurrency.MinTolerance())
	if err != nil {
		return -1
	}
	return rate
}

func (q QuotePair) Spread() int64 {
	return spread(q.BaseCurrency.MarketRate, q.QuoteCurrency.MarketRate, q.BaseCurrency.Min(), q.QuoteCurrency.Max())
}

func (q QuotePair) SpreadWithTolerance() int64 {
	return spread(q.BaseCurrency.MarketRate, q.QuoteCurrency.MarketRate, q.BaseCurrency.MinTolerance(), q.QuoteCurrency.MaxTolerance())
}

func (q QuotePair) Flip() QuotePair {
	return QuotePair{
		BaseCurrency:  q.QuoteCurrency,
		QuoteCurrency: q.BaseCurrency,
	}
}

func spread(srcRate, dstRate, sprdSrcRate, sprdDstRate uint64) int64 {
	mkt, err := conversions.Convert(1e8, srcRate, dstRate)
	if err != nil {
		return -1
	}

	tol, err := conversions.Convert(1e8, sprdSrcRate, sprdDstRate)
	if err != nil {
		return -1
	}

	// mkt should always be higher than tol
	if tol > mkt {
		return -1
	}

	return mkt - tol
}
