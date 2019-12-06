package pegnet

import (
	"database/sql"
	"time"

	"github.com/pegnet/pegnetd/fat/fat2"

	"github.com/Factom-Asset-Tokens/factom"
)

const createTableTransactionBatchHolding = `CREATE TABLE IF NOT EXISTS "pn_transaction_batch_holding" (
        "id"            	INTEGER PRIMARY KEY,
        "entry_hash"    	BLOB NOT NULL UNIQUE,
        "entry_data"    	BLOB NOT NULL,
        "height"        	INTEGER NOT NULL,
        "eblock_keymr"  	BLOB NOT NULL,
        "unix_timestamp"	INTEGER NOT NULL
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
                ("entry_hash", "entry_data", "height", "eblock_keymr", "unix_timestamp") VALUES
                (?, ?, ?, ?, ?)`)
	if err != nil {
		return -1, err
	}

	entryData, err := txBatch.Entry.MarshalBinary()
	if err != nil {
		return -1, err
	}
	res, err := stmt.Exec(txBatch.Entry.Hash[:], entryData, height, eblockKeyMR[:], txBatch.Entry.Timestamp.Unix())
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
	query := `SELECT "entry_data", "unix_timestamp" FROM "pn_transaction_batch_holding" WHERE "height" == ?;`
	rows, err := p.DB.Query(query, height)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txBatches []*fat2.TransactionBatch
	for rows.Next() {
		var entryData []byte
		var unix int64
		err = rows.Scan(&entryData, &unix)
		if err != nil {
			return nil, err
		}
		var entry factom.Entry
		err = entry.UnmarshalBinary(entryData)
		if err != nil {
			return nil, err
		}
		entry.Timestamp = time.Unix(unix, 0)
		txBatch, err := fat2.NewTransactionBatch(entry)
		if err != nil {
			continue // TODO: this should never happen?
		}
		txBatches = append(txBatches, txBatch)
	}
	return txBatches, nil
}

func (p *Pegnet) DoesTransactionExist(entryhash factom.Bytes32) (bool, error) {
	var found []byte
	query := `SELECT "entry_hash" FROM "pn_address_transactions" WHERE "entry_hash" == ?;`
	err := p.DB.QueryRow(query, entryhash[:]).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return found != nil && len(found) > 0, nil
}
