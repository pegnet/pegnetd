package pegnet

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnet/modules/opr"
	"github.com/pegnet/pegnetd/fat/fat2"
)

const createTableGrade = `CREATE TABLE IF NOT EXISTS "pn_grade" (
	"height" INTEGER PRIMARY KEY,
	"keymr" BLOB,
	"prevkeymr" BLOB,
	"eb_seq" INTEGER,
	"shorthashes" BLOB,
	"version" INTEGER,
	"cutoff" INTEGER,
	"count" INTEGER,
	
	UNIQUE("height")
);
`

const createTableWinners = `CREATE TABLE IF NOT EXISTS "pn_winners" (
	"height" INTEGER NOT NULL,
	"entryhash" BLOB,
	"oprhash" BLOB,
	"payout" INTEGER,
	"grade" REAL,
	"nonce" BLOB,
	"difficulty" BLOB, -- sqlite can't do uint64, stored as bigendian 8 bytes
	"position" INTEGER,
	"minerid" TEXT,
	"address" BLOB,
	UNIQUE("height", "position")
);`

const createTableRate = `CREATE TABLE IF NOT EXISTS "pn_rate" (
	"height" INTEGER NOT NULL,
	"token" TEXT,
	"value" INTEGER,
	"movingaverage" INTEGER DEFAULT 0,
	
	UNIQUE("height", "token")
);
`

func (p *Pegnet) insertRate(tx *sql.Tx, height uint32, tickerString string, rate uint64, mvAvg uint64) error {
	_, err := tx.Exec("INSERT INTO pn_rate (height, token, value, movingaverage) VALUES ($1, $2, $3, $4)", height, tickerString, rate, mvAvg)
	if err != nil {
		return err
	}
	return nil
}

const (
	PreviousWeight = 7
)

// ComputeMovingAverage is a weighted moving average
// ([Previous Avg * (nPoints - 1)] + Current Price) / nPoints
func ComputeMovingAverage(latest uint64, previous uint64, nPoints int) uint64 {
	if previous == 0 {
		return latest
	}

	l := new(big.Int).SetUint64(latest)
	p := new(big.Int).SetUint64(previous)
	w := big.NewInt(int64(nPoints) - 1)
	p = p.Mul(p, w) // Weight the previous
	p = p.Add(p, l) // Add the current price
	p = p.Quo(p, big.NewInt(int64(nPoints)))

	return p.Uint64()
}

// InsertRates adds all asset rates as rows, computing the rate for PEG if necessary
func (p *Pegnet) InsertRates(tx *sql.Tx, height uint32, rates []opr.AssetUint, pricePEG bool) error {
	// Rates are the spot prices for the asset for this block. We need to store
	// more than just the spot price, as we also need to include the average price.
	previousRates, err := p.SelectRecentRatesWithAvgs(height)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	for i := range rates {
		if rates[i].Name == "PEG" {
			continue
		}

		// Correct rates to use `pAsset`
		rates[i].Name = "p" + rates[i].Name

		// Calculate moving avg
		ticker := fat2.StringToTicker(rates[i].Name)
		mA := ComputeMovingAverage(rates[i].Value, previousRates[ticker].MovingAverage, PreviousWeight)

		err := p.insertRate(tx, height, rates[i].Name, rates[i].Value, mA)
		if err != nil {
			return err
		}
	}
	ratePEG := new(big.Int)
	if pricePEG {
		// PEG price = (total capitalization of all other assets) / (total supply of all other assets at height - 1)
		issuance, err := p.SelectIssuances()
		if err != nil {
			return err
		}
		totalCapitalization := new(big.Int)
		for _, r := range rates {
			if r.Name == "PEG" {
				continue
			}

			assetCapitalization := new(big.Int).Mul(new(big.Int).SetUint64(issuance[fat2.StringToTicker(r.Name)]), new(big.Int).SetUint64(r.Value))
			totalCapitalization.Add(totalCapitalization, assetCapitalization)
		}
		if issuance[fat2.PTickerPEG] == 0 {
			ratePEG.Set(totalCapitalization)
		} else {
			ratePEG.Div(totalCapitalization, new(big.Int).SetUint64(issuance[fat2.PTickerPEG]))
		}
	}

	mA := ComputeMovingAverage(ratePEG.Uint64(), previousRates[fat2.PTickerPEG].MovingAverage, PreviousWeight)
	err = p.insertRate(tx, height, fat2.PTickerPEG.String(), ratePEG.Uint64(), mA)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pegnet) InsertGradeBlock(tx *sql.Tx, eblock *factom.EBlock, graded grader.GradedBlock) error {
	data, err := json.Marshal(graded.WinnersShortHashes())
	if err != nil {
		return err
	}
	_, err = tx.Exec("INSERT INTO pn_grade (height, keymr, prevkeymr, eb_seq, shorthashes, version, cutoff, count) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		eblock.Height, eblock.KeyMR[:], eblock.PrevKeyMR[:], eblock.Sequence, data, graded.Version(), graded.Cutoff(), graded.Count())
	if err != nil {
		return err
	}

	// No winners? Then don't insert
	if len(graded.Winners()) > 0 {
		for _, o := range graded.Graded() {
			diff := make([]byte, 8)
			binary.BigEndian.PutUint64(diff, o.SelfReportedDifficulty)
			_, err = tx.Exec(`INSERT INTO pn_winners (height, entryhash, oprhash, payout, grade, nonce, difficulty, position, minerid, address) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
				eblock.Height, o.EntryHash, o.OPRHash, o.Payout(), o.Grade, o.Nonce, diff, o.Position(), o.OPR.GetID(), o.OPR.GetAddress())
			if err != nil {
				return fmt.Errorf("ht %d, pos %d :%s", eblock.Height, o.Position(), err)
			}
		}
	}

	return nil
}

func (p *Pegnet) SelectPreviousWinners(ctx context.Context, height uint32) ([]string, error) {
	var data []byte
	err := p.DB.QueryRowContext(ctx, "SELECT shorthashes FROM pn_grade WHERE height < $1 ORDER BY height DESC LIMIT 1", height).Scan(&data)
	if err != nil {
		return nil, err
	}

	var winners []string
	err = json.Unmarshal(data, &winners)
	if err != nil {
		return nil, err
	}

	return winners, nil
}

func (p *Pegnet) SelectPendingRates(ctx context.Context, tx *sql.Tx, height uint32) (map[fat2.PTicker]uint64, error) {
	rows, err := tx.Query("SELECT token, value FROM pn_rate WHERE height = $1", height)
	if err != nil {
		return nil, err
	}
	return _extractAssets(rows)
}

func (p *Pegnet) SelectPendingRatesWithAvgs(ctx context.Context, tx *sql.Tx, height uint32) (map[fat2.PTicker]Quote, error) {
	rows, err := tx.Query("SELECT token, value, movingaverage FROM pn_rate WHERE height = $1", height)
	if err != nil {
		return nil, err
	}
	return _extractAssetsWithAvg(rows)
}

// SelectRecentRatesWithAvgs selects the last rates from a height. It is possible to
// skip a block for rates, and sometimes the last recorded rates are needed.
func (p *Pegnet) SelectRecentRatesWithAvgs(height uint32) (map[fat2.PTicker]Quote, error) {
	rows, err := p.DB.Query("SELECT token, value, movingaverage FROM pn_rate WHERE height = (SELECT MAX(height) FROM pn_rate WHERE height <= $1)", height)
	if err != nil {
		return nil, err
	}
	return _extractAssetsWithAvg(rows)
}

func (p *Pegnet) SelectRates(ctx context.Context, height uint32) (map[fat2.PTicker]uint64, error) {
	rows, err := p.DB.Query("SELECT token, value FROM pn_rate WHERE height = $1", height)
	if err != nil {
		return nil, err
	}
	return _extractAssets(rows)
}

func (p *Pegnet) SelectRatesByKeyMR(ctx context.Context, keymr *factom.Bytes32) (map[fat2.PTicker]uint64, error) {
	rows, err := p.DB.Query("SELECT token, value FROM pn_rate WHERE height = (SELECT height FROM pn_grade WHERE keymr = $1)", keymr)
	if err != nil {
		return nil, err
	}
	return _extractAssets(rows)
}

func (p *Pegnet) SelectMostRecentRatesBeforeHeight(ctx context.Context, tx *sql.Tx, height uint32) (map[fat2.PTicker]uint64, uint32, error) {
	assets := make(map[fat2.PTicker]uint64)
	var rateHeight uint32
	queryString := `SELECT "token", "value", "height"
                    FROM "pn_rate" WHERE "height" = (
                        SELECT MAX("height")
                        FROM "pn_rate" WHERE "height" < ?
                    );`
	rows, err := tx.Query(queryString, height)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var tickerName string
		var rateValue uint64
		if err := rows.Scan(&tickerName, &rateValue, &rateHeight); err != nil {
			return nil, 0, err
		}
		assets[fat2.StringToTicker(tickerName)] = rateValue
	}
	if rows.Err() != nil {
		return nil, 0, err
	}
	return assets, rateHeight, nil
}

func _extractAssets(rows *sql.Rows) (map[fat2.PTicker]uint64, error) {
	defer rows.Close()
	assets := make(map[fat2.PTicker]uint64)
	for rows.Next() {
		var tickerName string
		var rateValue uint64
		if err := rows.Scan(&tickerName, &rateValue); err != nil {
			return nil, err
		}
		if ticker := fat2.StringToTicker(tickerName); ticker != fat2.PTickerInvalid {
			assets[ticker] = rateValue
		}
	}
	return assets, nil
}

type Quote struct {
	Price         uint64
	MovingAverage uint64
}

func (q Quote) Min() uint64 {
	if q.Price < q.MovingAverage {
		return q.Price
	}
	return q.MovingAverage
}

func (q Quote) Max() uint64 {
	if q.Price > q.MovingAverage {
		return q.Price
	}
	return q.MovingAverage
}

func _extractAssetsWithAvg(rows *sql.Rows) (map[fat2.PTicker]Quote, error) {
	defer rows.Close()
	assets := make(map[fat2.PTicker]Quote)
	for rows.Next() {
		var tickerName string
		var rateValue uint64
		var movingAvg uint64
		if err := rows.Scan(&tickerName, &rateValue, &movingAvg); err != nil {
			return nil, err
		}
		if ticker := fat2.StringToTicker(tickerName); ticker != fat2.PTickerInvalid {
			assets[ticker] = Quote{Price: rateValue, MovingAverage: movingAvg}
		}
	}
	return assets, nil
}
