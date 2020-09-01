package pegnet

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pegnet/pegnet/modules/conversions"
	"github.com/pegnet/pegnetd/fat/fat2"
)

const createTableStats = `CREATE TABLE IF NOT EXISTS "pn_stats" (
	"height" INTEGER,
	"data" BLOB
);
`

type Stats struct {
	Height          uint32
	Burns           uint64
	ConversionTotal uint64
	Supply          map[string]int64
	Volume          map[string]uint64
	VolumeIn        map[string]uint64
	VolumeOut       map[string]uint64
	VolumeTx        map[string]uint64
}

func NewStats(height uint32) *Stats {
	return &Stats{
		Height:    height,
		Supply:    make(map[string]int64),
		Volume:    make(map[string]uint64),
		VolumeIn:  make(map[string]uint64),
		VolumeOut: make(map[string]uint64),
		VolumeTx:  make(map[string]uint64),
	}
}

func (p *Pegnet) InsertStats(tx *sql.Tx, stats *Stats) error {
	// collect supply
	q := `SELECT
IFNULL(SUM(peg_balance),0),IFNULL(SUM(pusd_balance),0),IFNULL(SUM(peur_balance),0),IFNULL(SUM(pjpy_balance),0),IFNULL(SUM(pgbp_balance),0),IFNULL(SUM(pcad_balance),0),IFNULL(SUM(pchf_balance),0),IFNULL(SUM(pinr_balance),0),IFNULL(SUM(psgd_balance),0),IFNULL(SUM(pcny_balance),0),IFNULL(SUM(phkd_balance),0),IFNULL(SUM(pkrw_balance),0),IFNULL(SUM(pbrl_balance),0),IFNULL(SUM(pphp_balance),0),IFNULL(SUM(pmxn_balance),0),IFNULL(SUM(pxau_balance),0),IFNULL(SUM(pxag_balance),0),IFNULL(SUM(pxbt_balance),0),IFNULL(SUM(peth_balance),0),IFNULL(SUM(pltc_balance),0),IFNULL(SUM(prvn_balance),0),IFNULL(SUM(pxbc_balance),0),IFNULL(SUM(pfct_balance),0),IFNULL(SUM(pbnb_balance),0),IFNULL(SUM(pxlm_balance),0),IFNULL(SUM(pada_balance),0),IFNULL(SUM(pxmr_balance),0),IFNULL(SUM(pdash_balance),0),IFNULL(SUM(pzec_balance),0),IFNULL(SUM(pdcr_balance),0),IFNULL(SUM(paud_balance), 0),IFNULL(SUM(pnzd_balance), 0),IFNULL(SUM(psek_balance), 0),IFNULL(SUM(pnok_balance), 0),IFNULL(SUM(prub_balance), 0),IFNULL(SUM(pzar_balance), 0),IFNULL(SUM(ptry_balance), 0),IFNULL(SUM(peos_balance), 0),IFNULL(SUM(plink_balance), 0),IFNULL(SUM(patom_balance), 0),IFNULL(SUM(pbat_balance), 0),IFNULL(SUM(pxtz_balance), 0)
FROM pn_addresses
`

	sum := make([]int64, 42)
	err := tx.QueryRow(q).Scan(&sum[0], &sum[1], &sum[2], &sum[3], &sum[4], &sum[5], &sum[6], &sum[7], &sum[8], &sum[9], &sum[10], &sum[11], &sum[12], &sum[13], &sum[14], &sum[15], &sum[16], &sum[17], &sum[18], &sum[19], &sum[20], &sum[21], &sum[22], &sum[23], &sum[24], &sum[25], &sum[26], &sum[27], &sum[28], &sum[29], &sum[30], &sum[31], &sum[32], &sum[33], &sum[34], &sum[35], &sum[36], &sum[37], &sum[38], &sum[39], &sum[40], &sum[41])
	if err != nil {
		return err
	}

	// get running total
	prev, err := p.SelectPrevStats(tx, stats.Height)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	} else if prev != nil {
		stats.ConversionTotal = prev.ConversionTotal
	}

	for i, v := range sum {
		stats.Supply[fat2.PTicker(i+1).String()] = v
	}

	for k, v := range stats.Supply {
		if v <= 0 {
			delete(stats.Supply, k)
		}
	}

	rates, err := p.SelectPendingRates(nil, tx, stats.Height)
	if err != nil {
		return err
	}

	for k, v := range stats.VolumeOut {
		if k == "pUSD" {
			stats.ConversionTotal += v
		} else {
			ts := fat2.StringToTicker(k)
			pUSDEquiv, err := conversions.Convert(int64(v), rates[ts], rates[fat2.PTickerUSD])
			if err != nil {
				fmt.Println(rates)
				return err
			}
			stats.ConversionTotal += uint64(pUSDEquiv)
		}
	}

	js, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO pn_stats (height, data) VALUES ($1, $2)`, stats.Height, js)
	if err != nil {
		return err
	}

	return nil
}

func (p *Pegnet) SelectStats(ctx context.Context, height uint32) (*Stats, error) {
	var raw []byte
	err := p.DB.QueryRowContext(ctx, "SELECT data FROM pn_stats WHERE height = $1", height).Scan(&raw)
	if err != nil {
		return nil, err
	}

	stats := new(Stats)
	err = json.Unmarshal(raw, stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (p *Pegnet) SelectPrevStats(tx *sql.Tx, height uint32) (*Stats, error) {
	var raw []byte
	err := tx.QueryRow("SELECT data FROM pn_stats WHERE height <= $1 ORDER BY height DESC LIMIT 1", height).Scan(&raw)
	if err != nil {
		return nil, err
	}

	stats := new(Stats)
	err = json.Unmarshal(raw, stats)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
