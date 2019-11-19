package pegnet_test

import (
	"math/rand"
	"testing"

	. "github.com/pegnet/pegnetd/node/pegnet"
)

// TestPositiveSpread ensures the spread is never negative
func TestPositiveSpread(t *testing.T) {
	for i := 0; i < 100000; i++ {
		q := randomQuote()

		if q.SpreadWithTolerance() < 0 {
			t.Errorf("found negative spread")
		}

		if q.MakeBase(randomQuote()).SpreadWithTolerance() < 0 {
			t.Errorf("found negative pair spread")
		}

		pair := q.MakeQuote(randomQuote())
		if pair.SpreadWithTolerance() < 0 {
			t.Errorf("found negative pair spread: %d", pair.SpreadWithTolerance())
		}
	}

}

func TestSpreadVector(t *testing.T) {
	type Vect struct {
		Q         Quote
		ExpSpread int64
	}

	vecs := []Vect{
		{Q: Quote{MarketRate: 379480000, MovingAverage: 349605720}, ExpSpread: 26079480},
	}

	for _, v := range vecs {
		if v.Q.SpreadWithTolerance() != v.ExpSpread {
			t.Errorf("exp %d, got %d", v.ExpSpread, v.Q.SpreadWithTolerance())
		}
	}
}

func randomQuote() Quote {
	var q Quote
	q.MarketRate = rand.Uint64() % 100000 * 1e8
	q.MovingAverage = rand.Uint64() % 100000 * 1e8
	if q.MarketRate == 0 || q.MovingAverage == 0 {
		return randomQuote()
	}
	return q
}
