package pegnet

import (
	"database/sql"

	"github.com/pegnet/pegnetd/fat/fat2"
)

type HistoryAction int32

const (
	Transfer HistoryAction = iota
	Conversion
	Burn
)
const createTableTxHistory = `CREATE TABLE IF NOT EXISTS "pn_transaction_history" (
	"entry_hash"    BLOB NOT NULL,
	"height"        INTEGER NOT NULL, -- height the tx is in
	"txid"     		BLOB NOT NULL,
	"timestamp"		INTEGER NOT NULL,
	"executed"		INTEGER NOT NULL, -- -1 if failed, 0 if pending, height it was applied at otherwise

	PRIMARY KEY("entry_hash", "height"),
	FOREIGN KEY("height") REFERENCES "pn_grade"
);
CREATE INDEX IF NOT EXISTS "idx_transaction_history_entry_hash" ON "pn_transaction_history"("entry_hash");
CREATE INDEX IF NOT EXISTS "idx_transaction_history_txid" ON "pn_transaction_history"("txid");
CREATE INDEX IF NOT EXISTS "idx_transaction_history_timestamp" ON "pn_transaction_history"("timestamp");
`

const createTableTxHistoryAction = `CREATE TABLE IF NOT EXISTS "pn_transaction_history_action" (
	"entry_hash"	BLOB NOT NULL,
	"tx_index"		INTEGER NOT NULL,	-- the batch index
	"action_type"	INTEGER NOT NULL,
	"address"       BLOB NOT NULL,
	"asset"			STRING NOT NULL,
	"amount"		INTEGER NOT NULL,

	PRIMARY KEY("entry_hash", "tx_index"),
	FOREIGN KEY("entry_hash") REFERENCES "pn_transaction_history"
);
CREATE INDEX IF NOT EXISTS "idx_transaction_history_address" ON "pn_transaction_history"("address");`

const createTableTxHistoryActionData = `CREATE TABLE IF NOT EXISTS "pn_transaction_history_action_data" (
	"entry_hash"		BLOB NOT NULL,
	"tx_index"			INTEGER NOT NULL,	-- the batch index
	"tx_index_index" 	INTEGER NOT NULL,
	"to_address"		BLOB NOT NULL,		-- transfer recipient
	"to_asset"			STRING NOT NULL,	-- conversion/burn
	"to_amount"			INTEGER NOT NULL,
	
	PRIMARY KEY("entry_hash", "tx_index", "tx_index_index"),
	FOREIGN KEY("entry_hash") REFERENCES "pn_transaction_history"
);`

func (p *Pegnet) SetTransactionHistoryExecuted(tx *sql.Tx, txbatch *fat2.TransactionBatch, executed int64) error {
	stmt, err := tx.Prepare(`UPDATE "pn_transaction_history" SET executed = ? WHERE entry_hash = ?`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(executed, txbatch.Entry.Hash[:])
	if err != nil {
		return err
	}
	return nil
}

func (p *Pegnet) InsertTransactionHistoryTxBatch(tx *sql.Tx, txbatch *fat2.TransactionBatch, height uint32) error {
	stmt, err := tx.Prepare(`INSERT INTO "pn_transaction_history"
                (entry_hash, height, txid, timestamp, executed) VALUES
                (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(txbatch.Entry.Hash[:], height, "", txbatch.Entry.Timestamp.Unix(), 0)
	if err != nil {
		return err
	}

	actionStatement, err := tx.Prepare(`INSERT INTO "pn_transaction_history_action"
                (entry_hash, tx_index, action_type, address, asset, amount) VALUES
                (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	actionDataStatement, err := tx.Prepare(`INSERT INTO "pn_transaction_history_action_data"
                (entry_hash, tx_index, tx_index_index, to_address, to_asset, to_amount) VALUES
                (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	for index, action := range txbatch.Transactions {
		var typ HistoryAction
		if action.IsConversion() {
			typ = Conversion
		} else {
			typ = Transfer
		}

		_, err = actionStatement.Exec(txbatch.Entry.Hash[:], index, typ, action.Input.Address[:], action.Input.Type.String(), action.Input.Amount)
		if err != nil {
			return err
		}

		if action.IsConversion() {
			_, err = actionDataStatement.Exec(txbatch.Entry.Hash[:], index, 0, "", action.Conversion.String(), 0)
			if err != nil {
				return err
			}
		} else {
			for indexindex, data := range action.Transfers {
				_, err = actionDataStatement.Exec(txbatch.Entry.Hash[:], index, indexindex, data.Address[:], "", data.Amount)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
