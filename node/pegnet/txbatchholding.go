package pegnet

import (
	"database/sql"

	"github.com/pegnet/pegnetd/fat/fat2"

	"github.com/Factom-Asset-Tokens/factom"
)

const createTableTransactionBatchHolding = `CREATE TABLE IF NOT EXISTS "pn_transaction_batch_holding" (
        "id"            INTEGER PRIMARY KEY,
        "entry_hash"    BLOB NOT NULL UNIQUE,
        "entry_data"    BLOB NOT NULL,
        "height"        INTEGER NOT NULL,
        "eblock_keymr"  BLOB NOT NULL
);
`

// InsertTransactionBatchHolding inserts a row into "pn_transaction_batch_holding"
// that stores the entry data at its entry hash, along with the contextual information
// of the block height and eblock_keymr. If successful, the row id for the new row in
// "pn_transaction_batch_holding" is returned.
//
// Note: It is assumed that the entry stored has already been validated as a fat2
// transaction batch that contains at least one conversion. It is only put into
// holding to be executed against future asset exchange rates.
func (p *Pegnet) InsertTransactionBatchHolding(tx *sql.Tx, txBatch *fat2.TransactionBatch, height uint64, eblockKeyMR *factom.Bytes32) (int64, error) {
	stmt, err := tx.Prepare(`INSERT INTO "pn_transaction_batch_holding"
                ("entry_hash", "entry_data", "height", "eblock_keymr") VALUES
                (?, ?, ?, ?)`)
	if err != nil {
		return -1, err
	}

	entryData, err := txBatch.MarshalBinary()
	if err != nil {
		return -1, err
	}
	res, err := stmt.Exec(txBatch.Hash[:], entryData, height, eblockKeyMR)
	if err != nil {
		return -1, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return lastID, nil
}

// SelectTransactionBatchesInHoldingAtHeight selects all fat2.TransactionBatch entries
// that are in holding at the given height. It should be assumed that a TransactionBatch
// in the database has already returned nil for TransactionBatch.Validate() and also that
// TransactionBatch.HasConversions() returns true.
func (p *Pegnet) SelectTransactionBatchesInHoldingAtHeight(height uint64) ([]*fat2.TransactionBatch, error) {
	query := `SELECT "entry_data" FROM "pn_transaction_batch_holding" WHERE "height" == ?;`
	rows, err := p.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txBatches []*fat2.TransactionBatch
	for rows.Next() {
		var entryData []byte
		err = rows.Scan(&entryData)
		if err != nil {
			return nil, err
		}
		var entry factom.Entry
		err = entry.UnmarshalBinary(entryData)
		if err != nil {
			return nil, err
		}
		txBatch := fat2.NewTransactionBatch(entry)
		txBatches = append(txBatches, txBatch)
	}
	return txBatches, nil
}
