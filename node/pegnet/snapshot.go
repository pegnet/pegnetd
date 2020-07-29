package pegnet

import (
	"database/sql"
	"fmt"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"
)

const (
	// SnapshotRate is how often snapshots are taken and paid out on a block basis
	SnapshotRate = 144 // Default is 144, once a day
)

// SnapshotCurrent moves the current snapshot to the past, and updates the current snapshot.
// 1. Clear snapshot_past
// 2. Save snapshot_current to past
// 3. Delete snapshot_current
// 4. Save current balances to snapshot_current
func (p *Pegnet) SnapshotCurrent(tx QueryAble) error {
	// Move the current snapshot to snapshot_past
	//	We do a WHERE select otherwise SQLlite gives an error that we will prune the whole table
	_, err := tx.Exec(`DELETE FROM snapshot_past WHERE peg_balance >= 0`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO snapshot_past SELECT * FROM snapshot_current`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM snapshot_current WHERE peg_balance >= 0`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO snapshot_current SELECT * FROM pn_addresses`)
	if err != nil {
		return err
	}

	return nil
}

// SelectSnapshotBalances returns a map of all valid PTickers and their associated
// snapshot balances for the given address. The snapshot balance is the minimum of the past, and current
// snapshot.
// You must provide the table to query on
func (Pegnet) SelectSnapshotBalances(tx QueryAble) ([]BalancesPair, error) {
	// This query merges all addresses that exist in both snapshots. The balance
	// in the column is the minimum balance of the 2 snapshots. If the
	// address does not exist in either column, it will not be present in the
	// resulting query.
	query := fmt.Sprintf(`SELECT %s
		FROM snapshot_past as sn_past
       		INNER JOIN snapshot_current as sn_current
		WHERE sn_past.address = sn_current.address;`, snapshotMinSelectCols)
	rows, err := tx.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var res []BalancesPair
	for rows.Next() {
		var bp BalancesPair
		bp.Balances = make([]uint64, int(fat2.PTickerMax)+1)

		var id int
		var address []byte
		err = rows.Scan(
			&id,
			&address,
			&bp.Balances[fat2.PTickerPEG],
			&bp.Balances[fat2.PTickerUSD],
			&bp.Balances[fat2.PTickerEUR],
			&bp.Balances[fat2.PTickerJPY],
			&bp.Balances[fat2.PTickerGBP],
			&bp.Balances[fat2.PTickerCAD],
			&bp.Balances[fat2.PTickerCHF],
			&bp.Balances[fat2.PTickerINR],
			&bp.Balances[fat2.PTickerSGD],
			&bp.Balances[fat2.PTickerCNY],
			&bp.Balances[fat2.PTickerHKD],
			&bp.Balances[fat2.PTickerKRW],
			&bp.Balances[fat2.PTickerBRL],
			&bp.Balances[fat2.PTickerPHP],
			&bp.Balances[fat2.PTickerMXN],
			&bp.Balances[fat2.PTickerXAU],
			&bp.Balances[fat2.PTickerXAG],
			&bp.Balances[fat2.PTickerXBT],
			&bp.Balances[fat2.PTickerETH],
			&bp.Balances[fat2.PTickerLTC],
			&bp.Balances[fat2.PTickerRVN],
			&bp.Balances[fat2.PTickerXBC],
			&bp.Balances[fat2.PTickerFCT],
			&bp.Balances[fat2.PTickerBNB],
			&bp.Balances[fat2.PTickerXLM],
			&bp.Balances[fat2.PTickerADA],
			&bp.Balances[fat2.PTickerXMR],
			&bp.Balances[fat2.PTickerDASH],
			&bp.Balances[fat2.PTickerZEC],
			&bp.Balances[fat2.PTickerDCR],
			// V4 Additions
			&bp.Balances[fat2.PTickerAUD],
			&bp.Balances[fat2.PTickerNZD],
			&bp.Balances[fat2.PTickerSEK],
			&bp.Balances[fat2.PTickerNOK],
			&bp.Balances[fat2.PTickerRUB],
			&bp.Balances[fat2.PTickerZAR],
			&bp.Balances[fat2.PTickerTRY],
			&bp.Balances[fat2.PTickerEOS],
			&bp.Balances[fat2.PTickerLINK],
			&bp.Balances[fat2.PTickerATOM],
			&bp.Balances[fat2.PTickerBAT],
			&bp.Balances[fat2.PTickerXTZ],
			// v5 Additions
			&bp.Balances[fat2.PTickerHBAR],
			&bp.Balances[fat2.PTickerNEO],
			&bp.Balances[fat2.PTickerCRO],
			&bp.Balances[fat2.PTickerETC],
			&bp.Balances[fat2.PTickerONT],
			&bp.Balances[fat2.PTickerDOGE],
			&bp.Balances[fat2.PTickerVET],
			&bp.Balances[fat2.PTickerHT],
			&bp.Balances[fat2.PTickerALGO],
			&bp.Balances[fat2.PTickerDGB],
			&bp.Balances[fat2.PTickerAED],
			&bp.Balances[fat2.PTickerARS],
			&bp.Balances[fat2.PTickerTWD],
			&bp.Balances[fat2.PTickerRWF],
			&bp.Balances[fat2.PTickerKES],
			&bp.Balances[fat2.PTickerUGX],
			&bp.Balances[fat2.PTickerTZS],
			&bp.Balances[fat2.PTickerBIF],
			&bp.Balances[fat2.PTickerETB],
			&bp.Balances[fat2.PTickerNGN],
		)

		if err != nil {
			return nil, err
		}

		var fa factom.FAAddress
		copy(fa[:], address)
		bp.Address = &fa

		res = append(res, bp)
	}
	return res, nil
}
