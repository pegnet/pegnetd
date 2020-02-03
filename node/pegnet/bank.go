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
// The bank has the following data structure at each height. It can be
// used to control the amount of PEG allowed to be converted into at each
// pegnet block.
//
// There is extra information provided for informational purposes. The only
// true value needed here is likely the BankAmount and possible the BankUsed.
// The PEGRequested shows the demand for PEG at a high level.
type BankEntry struct {
	// Each height will have a bank entry.
	Height int32
	// BankAmount is the total amount of PEG allowed to be converted into
	// for the given height.
	BankAmount int64 // units are PEGtoshi
	// BankUsed is some additional information to detail how much of the bank
	// was consumed.
	BankUsed int64 // units are PEGtoshi
	// PEGRequested is the total amount of PEG requested
	PEGRequested int64 // units are PEGtoshi
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

// InsertBankAmount does not fill in the bank_used and total_requested with
// legit values. It leaves a -1 to indicate that needs to be filled.
func (p Pegnet) InsertBankAmount(q QueryAble, height int32, bankAmount int64) error {
	if q == nil {
		q = p.DB // nil defaults to db
	}

	query := fmt.Sprintf(`INSERT INTO pn_bank("height", "bank_amount", "bank_used", "total_requested") VALUES(?, ?, -1, -1);`)
	res, err := q.Exec(query, height, bankAmount)
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

// UpdateBankEntry updates the bank_used and total_requested values
func (p Pegnet) UpdateBankEntry(q QueryAble, height int32, bankUsed, pegRequested int64) error {
	if q == nil {
		q = p.DB // nil defaults to db
	}

	query := fmt.Sprintf(`
	UPDATE pn_bank
		SET bank_used = ?,
			total_requested = ?
		WHERE height = ?;
`)

	res, err := q.Exec(query, bankUsed, pegRequested, height)
	if err != nil {
		return err
	}
	if aff, err := res.RowsAffected(); err != nil {
		return err
	} else if aff != 1 {
		return fmt.Errorf("bank entry not updated")
	}
	return nil
}
