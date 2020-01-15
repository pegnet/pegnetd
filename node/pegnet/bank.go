package pegnet

import (
	"database/sql"
	"fmt"

	"github.com/pegnet/pegnet/modules/conversions"
)

// BankEntry is for managing the PEG Bank.
// The bank is the amount of PEG allowed to be converted into every block.
// This table should be used for bank management purposes.
//
// Originally the bank was thought of in order to allow conversions out of PEG
// to positively impact the bank, allowing more conversions in.
// This would require the supply to be under the supply curve, which we are
// no where close too.
//
// So now the bank has a different control. The bank will be controlled
// by other mechanisms. Each height will have the following struct data stored
//
// There is extra information provided for informational purposes. The only
// true value needed here is likely the BankAmount and possible the BankUsed.
// The PEGRequested shows the demand for PEG at a high level.
type BankEntry struct {
	// Each height will have a bank entry.
	Height int32
	// BankAmount is the total amount of PEG allowed to be converted into
	// for the given height.
	BankAmount int64 // units are PEG
	// BankUsed is some additional information to detail how much of the bank
	// was consumed.
	BankUsed int64 // units are PEG
	// PEGRequested is the total amount of PEG requested
	PEGRequested int64 // units are PEG
}

const createTableBank = `CREATE TABLE IF NOT EXISTS "pn_bank" (
	"height" INTEGER PRIMARY KEY,
	"bank_amount" INTEGER NOT NULL DEFAULT 0,
	"bank_used" INTEGER NOT NULL DEFAULT 0,
	"total_requested" INTEGER NOT NULL DEFAULT 0,
	
	UNIQUE("height")
);
`

const (
	// Bank Properties and constants

	// BankBaseAmount is the minimum the bank size will ever be.
	// Any bank growth will be ontop of this value.
	BankBaseAmount = conversions.PerBlock

	// BankGrowth
	BankGrowthAmount = 500 * 1e8
	BankDecayAmount  = 500 * 1e8

	// BankMaxLimit is the maximum size of the Bank for any given block
	// 5 * 1 million = 5 million PEG
	million      = 1e6
	BankMaxLimit = 5 * million * 1e8
)

// CreateTableBank is used to expose this table for unit tests
func (p *Pegnet) CreateTableBank() error {
	_, err := p.DB.Exec(createTableBank)
	if err != nil {
		return err
	}
	return nil
}

func (p Pegnet) SelectBankEntry(q QueryAble, height int32) (entry BankEntry, err error) {
	if q == nil {
		q = p.DB // nil defaults to db
	}

	query := fmt.Sprintf(`SELECT "height", "bank_amount", "bank_used", "total_requested" FROM pn_bank WHERE height = ?;`)
	err = q.QueryRow(query, height).Scan(&entry.Height, &entry.BankAmount, &entry.BankUsed, &entry.PEGRequested)
	if err == sql.ErrNoRows {
		entry = BankEntry{Height: -1, BankAmount: -1}
		err = nil
	}
	return
}

// InsertBankEntry
func (p Pegnet) InsertBankEntry(q QueryAble, height int32, bankAmount, bankUsed, pegRequested int64) error {
	if q == nil {
		q = p.DB // nil defaults to db
	}

	query := fmt.Sprintf(`INSERT INTO pn_bank("height", "bank_amount", "bank_used", "total_requested") VALUES(?, ?, ?, ?);`)
	res, err := q.Exec(query, height, bankAmount, bankUsed, pegRequested)
	if err != nil {
		return err
	}
	if aff, err := res.RowsAffected(); err != nil {
		return err
	} else if aff != 1 {
		return fmt.Errorf("bank entry not added")
	}
	return nil
}
