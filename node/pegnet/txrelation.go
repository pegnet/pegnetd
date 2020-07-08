package pegnet

import (
	"database/sql"

	"github.com/Factom-Asset-Tokens/factom"
)

// createTableTransactions is a SQL string that creates the
// "pn_address_transactions" table.
//
// The "pn_address_transactions" table has a foreign key reference to the
// "pn_addresses" table, which must exist first.
//
// If the transaction is a conversion, both "to" and "conversion" are set to true
const createTableTransactions = `CREATE TABLE IF NOT EXISTS "pn_address_transactions" (
        "entry_hash"    BLOB NOT NULL,
        "address"       BLOB NOT NULL,
        "tx_index"      INTEGER NOT NULL,
        "to"            BOOL NOT NULL,
        "conversion"    BOOL NOT NULL,

        PRIMARY KEY("entry_hash", "address"),

        FOREIGN KEY("address") REFERENCES "pn_addresses"
);
CREATE INDEX IF NOT EXISTS "idx_address_transactions_address" ON "pn_address_transactions"("address");
`

// InsertTransactionRelation inserts a row into "pnaddress_transactions" relating
// the adrID with the entryHash with the given transaction direction, to. If
// successful, the row id for the new row in "pn_address_transactions" is
// returned.
//
// If isConversion is true, the to field will automatically be set to true.
func (p *Pegnet) InsertTransactionRelation(tx *sql.Tx, adr factom.FAAddress, entryHash *factom.Bytes32, txIndex uint64, to bool, isConversion bool) (int64, error) {
	// If an address is the sender and the receiver, then we only record the sender side, not the receiver.
	// This is some loss of information, but for the use case of getting all related transactions,
	// it is fine.
	stmt, err := tx.Prepare(`INSERT INTO "pn_address_transactions"
                ("entry_hash", "address", "tx_index", "to", "conversion") VALUES
                (?, ?, ?, ?, ?) ON CONFLICT DO NOTHING`)
	if err != nil {
		return -1, err
	}
	res, err := stmt.Exec(entryHash[:], adr[:], txIndex, to || isConversion, isConversion)
	if err != nil {
		return -1, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return lastID, nil
}

// IsReplayTransaction returns true if there exist any transaction relations in the
// "pn_address_transactions" table.
func (p *Pegnet) IsReplayTransaction(tx *sql.Tx, entryHash *factom.Bytes32) (bool, error) {
	rows, err := tx.Query(`SELECT * FROM "pn_address_transactions" WHERE "entry_hash" = ?;`, entryHash[:])
	if err != nil {
		return false, err
	}
	defer rows.Close()
	err = rows.Err()
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	// If there is any result, then we know the transaction has been executed before and thus a replay.
	return rows.Next(), nil
}
