package pegnet

import (
	"database/sql"
	"encoding/json"

	"github.com/pegnet/pegnetd/fat/fat2"
)

const createTableStats = `CREATE TABLE IF NOT EXISTS "pn_stats" (
	"height" INTEGER,
	"data" BLOB
);
`

type Stats struct {
	Height uint32
	Burns  uint64
	Supply map[string]int64
	Volume map[string]uint64
}

func NewStats(height uint32) *Stats {
	return &Stats{
		Height: height,
		Supply: make(map[string]int64),
		Volume: make(map[string]uint64),
	}
}

func (p *Pegnet) InsertStats(tx *sql.Tx, stats *Stats) error {

	// collect supply
	q := `SELECT
SUM(peg_balance),SUM(pusd_balance),SUM(peur_balance),SUM(pjpy_balance),SUM(pgbp_balance),SUM(pcad_balance),SUM(pchf_balance),SUM(pinr_balance),SUM(psgd_balance),SUM(pcny_balance),SUM(phkd_balance),SUM(pkrw_balance),SUM(pbrl_balance),SUM(pphp_balance),SUM(pmxn_balance),SUM(pxau_balance),SUM(pxag_balance),SUM(pxbt_balance),SUM(peth_balance),SUM(pltc_balance),SUM(prvn_balance),SUM(pxbc_balance),SUM(pfct_balance),SUM(pbnb_balance),SUM(pxlm_balance),SUM(pada_balance),SUM(pxmr_balance),SUM(pdas_balance),SUM(pzec_balance),SUM(pdcr_balance)
FROM pn_addresses
`
	sum := make([]int64, 30)
	err := tx.QueryRow(q).Scan(&sum[0], &sum[1], &sum[2], &sum[3], &sum[4], &sum[5], &sum[6], &sum[7], &sum[8], &sum[9], &sum[10], &sum[11], &sum[12], &sum[13], &sum[14], &sum[15], &sum[16], &sum[17], &sum[18], &sum[19], &sum[20], &sum[21], &sum[22], &sum[23], &sum[24], &sum[25], &sum[26], &sum[27], &sum[28], &sum[29])
	if err != nil {
		return err
	}

	for i, v := range sum {
		stats.Supply[fat2.PTicker(i+1).String()] = v
	}

	for k, v := range stats.Supply {
		if v <= 0 {
			delete(stats.Supply, k)
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
