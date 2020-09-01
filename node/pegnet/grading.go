package pegnet

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

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
	
	UNIQUE("height", "token")
);
`

func (p *Pegnet) insertRate(tx *sql.Tx, height uint32, tickerString string, rate uint64) error {
	_, err := tx.Exec("INSERT INTO pn_rate (height, token, value) VALUES ($1, $2, $3)", height, tickerString, rate)
	if err != nil {
		return err
	}
	return nil
}

type PEGPricingPhase int

const (
	_                  PEGPricingPhase = iota
	PEGPriceIsZero                     // PEG == 0
	PEGPriceIsEquation                 // PEG == MarketCap / Peg Supply
	PEGPriceIsFloating                 // PEG == ExchRate
)

const (
	// Prefix for pAsset prices on exchanges reported by miners
	PAssetExchangePrefix = "exch_"
)

// InsertRates adds all asset rates as rows, computing the rate for PEG if necessary
func (p *Pegnet) InsertRates(tx *sql.Tx, height uint32, rates []opr.AssetUint, phase PEGPricingPhase) error {
	if phase == 0 {
		return fmt.Errorf("undefined PEG phase")
	}

	ratePEG := new(big.Int)
	for i := range rates {
		if rates[i].Name == "PEG" {
			ratePEG.SetUint64(rates[i].Value)
			continue
		}

		// Correct rates to use `pAsset`
		rates[i].Name = "p" + rates[i].Name

		err := p.insertRate(tx, height, rates[i].Name, rates[i].Value)
		if err != nil {
			return err
		}
	}

	// Now to insert the PEG rate. All other rates are set above.
	// The PEG rate depends on what activation phase we are in. There are
	// multiple ways to set the PEG price.
	switch phase {
	case PEGPriceIsZero: // PEG Price is 0
		ratePEG.SetUint64(0)
	case PEGPriceIsEquation: // Market Cap Equation
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
			// If there are no PEGs in the system, PEGs have no value (divide-by-zero)
			// At least one block will have to be mined in order for PEGs to attain a value
			ratePEG.SetUint64(0)
		} else {
			ratePEG.Div(totalCapitalization, new(big.Int).SetUint64(issuance[fat2.PTickerPEG]))
		}
	case PEGPriceIsFloating: // Rate in opr is the rate
	}

	err := p.insertRate(tx, height, fat2.PTickerPEG.String(), ratePEG.Uint64())
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

// SelectReferenceRates returns the pAsset rates on the external market that
// are reported by the miners.
//
// So pUSD == $1, but the Reference token (pUSD) can be trading at $0.90
// This rate call will return $0.90
func (p *Pegnet) SelectReferenceRates(ctx context.Context, tx QueryAble, height uint32) (map[fat2.PTicker]uint64, error) {
	if tx == nil {
		tx = p.DB
	}
	rows, err := tx.Query("SELECT token, value FROM pn_rate WHERE height = $1", height)
	if err != nil {
		return nil, err
	}
	return _extractAssetsWithPrefix(rows, PAssetExchangePrefix)
}

func (p *Pegnet) SelectPendingRates(ctx context.Context, tx *sql.Tx, height uint32) (map[fat2.PTicker]uint64, error) {
	rows, err := tx.Query("SELECT token, value FROM pn_rate WHERE height = $1", height)
	if err != nil {
		return nil, err
	}
	return _extractAssets(rows)
}

func (p *Pegnet) SelectRates(ctx context.Context, height uint32) (map[fat2.PTicker]uint64, error) {
	rows, err := p.DB.Query("SELECT token, value FROM pn_rate WHERE height = $1", height)
	if err != nil {
		return nil, err
	}
	return _extractAssets(rows)
}
func (p *Pegnet) SelectRecentRates(ctx context.Context, height uint32) (map[string]uint64, error) {
	rows, err := p.DB.Query("SELECT token, value FROM pn_rate WHERE height = (SELECT MAX(height) FROM pn_rate WHERE height <= $1)", height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := make(map[string]uint64)
	for rows.Next() {
		var tickerName string
		var rateValue uint64
		if err := rows.Scan(&tickerName, &rateValue); err != nil {
			return nil, err
		}
		assets[tickerName] = rateValue
	}
	return assets, nil
}

func (p *Pegnet) SelectRatesByKeyMR(ctx context.Context, keymr *factom.Bytes32) (map[fat2.PTicker]uint64, error) {
	rows, err := p.DB.Query("SELECT token, value FROM pn_rate WHERE height = (SELECT height FROM pn_grade WHERE keymr = $1)", keymr)
	if err != nil {
		return nil, err
	}
	return _extractAssets(rows)
}

func (p *Pegnet) SelectMostRecentRatesBeforeHeight(ctx context.Context, tx QueryAble, height uint32) (map[fat2.PTicker]uint64, uint32, error) {
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
	return _extractAssetsWithPrefix(rows, "")
}

func _extractAssetsWithPrefix(rows *sql.Rows, prefix string) (map[fat2.PTicker]uint64, error) {
	defer rows.Close()
	assets := make(map[fat2.PTicker]uint64)
	for rows.Next() {
		var tickerName string
		var rateValue uint64
		if err := rows.Scan(&tickerName, &rateValue); err != nil {
			return nil, err
		}
		if !strings.HasPrefix(tickerName, prefix) {
			continue
		}
		trimmed := strings.TrimPrefix(tickerName, prefix)
		if ticker := fat2.StringToTicker(trimmed); ticker != fat2.PTickerInvalid {
			assets[ticker] = rateValue
		}
	}
	return assets, nil
}
