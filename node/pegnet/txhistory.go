package pegnet

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnet/modules/grader"
	"github.com/pegnet/pegnetd/fat/fat2"
)

// HistoryTransaction is a flattened entry of the history table structure.
// It contains several actions: transfers, conversions, coinbases, and fct burns
type HistoryTransaction struct {
	Hash      *factom.Bytes32 `json:"hash"`
	TxID      string          `json:"txid"` // [TxIndex]-[BatchHash]
	Height    int64           `json:"height"`
	Timestamp time.Time       `json:"timestamp"`
	Executed  int32           `json:"executed"`
	TxIndex   int             `json:"txindex"`
	TxAction  HistoryAction   `json:"txaction"`

	FromAddress *factom.FAAddress          `json:"fromaddress"`
	FromAsset   string                     `json:"fromasset"`
	FromAmount  int64                      `json:"fromamount"`
	ToAsset     string                     `json:"toasset,omitempty"`
	ToAmount    int64                      `json:"toamount,omitempty"`
	Outputs     []HistoryTransactionOutput `json:"outputs,omitempty"`
}

// HistoryTransactionOutput is an entry of a transfer's outputs
type HistoryTransactionOutput struct {
	Address factom.FAAddress `json:"address"`
	Amount  int64            `json:"amount"`
}

// in the context of tables, `history_txbatch` is the table that holds the unique reference hash
// and `transaction` is the table that holds the actions associated with that unique reference hash
// `lookup` is an outside reference that indexes the addresses involved in the actions
//
// associations are:
// 	* history_txbach : history_transaction is `1:n`
// 	* history_transaction : lookup is `1:n`
//	* lookup : (transaction.outputs + transaction.inputs) is `1:n` (unique addresses only)
const createTableTxHistoryBatch = `CREATE TABLE IF NOT EXISTS "pn_history_txbatch" (
	"history_id"	INTEGER PRIMARY KEY,
	"entry_hash"    BLOB NOT NULL,
	"height"        INTEGER NOT NULL, -- height the tx is in
	"blockorder"	INTEGER NOT NULL,
	"timestamp"		INTEGER NOT NULL,
	"executed"		INTEGER NOT NULL, -- -1 if failed, 0 if pending, height it was applied at otherwise

	UNIQUE("entry_hash", "height")
);
CREATE INDEX IF NOT EXISTS "idx_history_txbatch_entry_hash" ON "pn_history_txbatch"("entry_hash");
CREATE INDEX IF NOT EXISTS "idx_history_txbatch_timestamp" ON "pn_history_txbatch"("timestamp");
CREATE INDEX IF NOT EXISTS "idx_history_txbatch_height" ON "pn_history_txbatch"("height");
`

const createTableTxHistoryTx = `CREATE TABLE IF NOT EXISTS "pn_history_transaction" (
	"entry_hash"	BLOB NOT NULL,
	"tx_index"		INTEGER NOT NULL,	-- the batch index
	"action_type"	INTEGER NOT NULL,
	"from_address"  BLOB NOT NULL,
	"from_asset"	STRING NOT NULL,
	"from_amount"	INTEGER NOT NULL,
	"to_asset"		STRING NOT NULL,	-- used for NOT transfers
	"to_amount"		INTEGER NOT NULL,	-- used for NOT transfers
	"outputs"		BLOB NOT NULL,		-- used for transfers only

	PRIMARY KEY("entry_hash", "tx_index"),
	FOREIGN KEY("entry_hash") REFERENCES "pn_history_txbatch"
);
CREATE INDEX IF NOT EXISTS "idx_history_transaction_entry_hash" ON "pn_history_transaction"("entry_hash");
CREATE INDEX IF NOT EXISTS "idx_history_transaction_tx_index" ON "pn_history_transaction"("tx_index");
`

const createTableTxHistoryLookup = `CREATE TABLE IF NOT EXISTS "pn_history_lookup" (
	"entry_hash"	INTEGER NOT NULL,
	"tx_index"		INTEGER NOT NULL,
	"address"		BLOB NOT NULL,

	PRIMARY KEY("entry_hash", "tx_index", "address"),
	FOREIGN KEY("entry_hash", "tx_index") REFERENCES "pn_history_transaction"
);
CREATE INDEX IF NOT EXISTS "idx_history_lookup_address" ON "pn_history_lookup"("address");
CREATE INDEX IF NOT EXISTS "idx_history_lookup_entry_index" ON "pn_history_lookup"("entry_hash", "tx_index");`

// only add a lookup reference if one doesn't already exist
const insertLookupQuery = `INSERT INTO pn_history_lookup (entry_hash, tx_index, address) VALUES (?, ?, ?) ON CONFLICT DO NOTHING;`

func (p *Pegnet) historySelectHelper(field string, data interface{}, options HistoryQueryOptions) ([]HistoryTransaction, int, error) {
	countQuery, dataQuery, err := historyQueryBuilder(field, options)
	if err != nil {
		return nil, 0, err
	}

	var count int
	err = p.DB.QueryRow(countQuery, data).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return nil, 0, nil
	}

	if options.Offset > count {
		return nil, 0, fmt.Errorf("offset too big")
	}

	rows, err := p.DB.Query(dataQuery, data)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	actions, err := turnRowsIntoHistoryTransactions(rows)
	return actions, count, err
}

// SelectTransactionHistoryActionsByHash returns the specified amount of transactions based on the hash.
// Hash can be an entry hash from the opr and transaction chains, or a transaction hash from an fblock.
func (p *Pegnet) SelectTransactionHistoryActionsByHash(hash *factom.Bytes32, options HistoryQueryOptions) ([]HistoryTransaction, int, error) {
	return p.historySelectHelper("entry_hash", hash[:], options)
}

// SelectTransactionHistoryActionsByAddress uses the lookup table to retrieve all transactions that have
// the specified address in either inputs or outputs
func (p *Pegnet) SelectTransactionHistoryActionsByAddress(addr *factom.FAAddress, options HistoryQueryOptions) ([]HistoryTransaction, int, error) {
	return p.historySelectHelper("address", addr[:], options)
}

// SelectTransactionHistoryActionsByTxID uses the lookup table to retrieve all transactions that have
// the specified txid. A TxID is an entryhash + a transaction index
func (p *Pegnet) SelectTransactionHistoryActionsByTxID(hash *factom.Bytes32, options HistoryQueryOptions) ([]HistoryTransaction, int, error) {
	return p.historySelectHelper("entry_hash", hash[:], options)
}

// SelectTransactionHistoryActionsByHeight returns all transactions that were **entered** at the specified height.
func (p *Pegnet) SelectTransactionHistoryActionsByHeight(height uint32, options HistoryQueryOptions) ([]HistoryTransaction, int, error) {
	return p.historySelectHelper("height", height, options)
}

// SelectTransactionHistoryStatus returns the status of a transaction:
// `-1` for a failed transaction, `0` for a pending transactions,
// `height` for the block in which it was applied otherwise
func (p *Pegnet) SelectTransactionHistoryStatus(hash *factom.Bytes32) (uint32, uint32, error) {
	var height, executed uint32
	err := p.DB.QueryRow("SELECT height, executed FROM pn_history_txbatch WHERE entry_hash = ?", hash[:]).Scan(&height, &executed)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return height, executed, nil
}

// SetTransactionHistoryExecuted updates a transaction's executed status
func (p *Pegnet) SetTransactionHistoryExecuted(tx *sql.Tx, txbatch *fat2.TransactionBatch, executed int64) error {
	stmt, err := tx.Prepare(`UPDATE "pn_history_txbatch" SET executed = ? WHERE entry_hash = ?`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(executed, txbatch.Entry.Hash[:])
	if err != nil {
		return err
	}
	return nil
}

// SetTransactionHistoryConvertedAmount updates a conversion with the actual conversion value.
// This is done in the same SQL Transaction as updating its executed status
func (p *Pegnet) SetTransactionHistoryConvertedAmount(tx *sql.Tx, txbatch *fat2.TransactionBatch, index int, amount int64) error {
	stmt, err := tx.Prepare(`UPDATE "pn_history_transaction" SET to_amount = ? WHERE entry_hash = ? AND tx_index = ?`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(amount, txbatch.Entry.Hash[:], index)
	if err != nil {
		return err
	}
	return nil
}

// InsertTransactionHistoryTxBatch inserts a transaction from the transaction chain into the history system
func (p *Pegnet) InsertTransactionHistoryTxBatch(tx *sql.Tx, blockorder int, txbatch *fat2.TransactionBatch, height uint32) error {
	stmt, err := tx.Prepare(`INSERT INTO "pn_history_txbatch"
                (entry_hash, height, blockorder, timestamp, executed) VALUES
                (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(txbatch.Entry.Hash[:], height, blockorder, txbatch.Entry.Timestamp.Unix(), 0)
	if err != nil {
		return err
	}

	txStatement, err := tx.Prepare(`INSERT INTO "pn_history_transaction"
                (entry_hash, tx_index, action_type, from_address, from_asset, from_amount, to_asset, to_amount, outputs) VALUES
                (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	lookup, err := tx.Prepare(insertLookupQuery)
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

		if _, err = lookup.Exec(txbatch.Entry.Hash[:], index, action.Input.Address[:]); err != nil {
			return err
		}

		if action.IsConversion() {
			_, err = txStatement.Exec(txbatch.Entry.Hash[:], index, typ,
				action.Input.Address[:], action.Input.Type.String(), action.Input.Amount, // from
				action.Conversion.String(), 0, "") // to
			if err != nil {
				return err
			}
		} else {
			// json encode the outputs
			outputs := make([]HistoryTransactionOutput, len(action.Transfers))
			for i, transfer := range action.Transfers {
				outputs[i] = HistoryTransactionOutput{Address: transfer.Address, Amount: int64(transfer.Amount)}
				if _, err = lookup.Exec(txbatch.Entry.Hash[:], index, transfer.Address[:]); err != nil {
					return err
				}
			}
			var outputData []byte
			if outputData, err = json.Marshal(outputs); err != nil {
				return err
			}

			if _, err = txStatement.Exec(txbatch.Entry.Hash[:], index, typ,
				action.Input.Address[:], action.Input.Type.String(), action.Input.Amount,
				"", 0, outputData); err != nil {
				return err
			}
		}
	}

	return nil
}

// InsertFCTBurn inserts a payout for an FCT burn into the system.
// Note that from_asset and to_asset are hardcoded
func (p *Pegnet) InsertFCTBurn(tx *sql.Tx, fBlockHash *factom.Bytes32, burn *factom.FactoidTransaction, height uint32) error {
	stmt, err := tx.Prepare(`INSERT INTO "pn_history_txbatch"
                (entry_hash, height, blockorder, timestamp, executed) VALUES
                (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	lookup, err := tx.Prepare(insertLookupQuery)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(burn.TransactionID[:], height, -1, burn.FactoidTransactionHeader.Timestamp.Unix(), height)
	if err != nil {
		return err
	}

	burnStatement, err := tx.Prepare(`INSERT INTO "pn_history_transaction"
                (entry_hash, tx_index, action_type, from_address, from_asset, from_amount, to_asset, to_amount, outputs) VALUES
                (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	if _, err = burnStatement.Exec(burn.TransactionID[:], 0, FCTBurn, burn.FCTInputs[0].Address[:], "FCT", burn.FCTInputs[0].Amount, "pFCT", burn.FCTInputs[0].Amount, ""); err != nil {
		return err
	}

	if _, err = lookup.Exec(burn.TransactionID[:], 0, burn.FCTInputs[0].Address[:]); err != nil {
		return err
	}

	return nil
}

// InsertCoinbase inserts the payouts from mining into the history system.
// There is one transaction per winning OPR, with the entry hash pointing to that specific opr
func (p *Pegnet) InsertCoinbase(tx *sql.Tx, winner *grader.GradingOPR, addr []byte, timestamp time.Time) error {
	stmt, err := tx.Prepare(`INSERT INTO "pn_history_txbatch"
                (entry_hash, height, blockorder, timestamp, executed) VALUES
                (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	lookup, err := tx.Prepare(insertLookupQuery)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(winner.EntryHash, winner.OPR.GetHeight(), 0, timestamp.Unix(), winner.OPR.GetHeight())
	if err != nil {
		return err
	}

	coinbaseStatement, err := tx.Prepare(`INSERT INTO "pn_history_transaction"
                (entry_hash, tx_index, action_type, from_address, from_asset, from_amount, to_asset, to_amount, outputs) VALUES
                (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}

	_, err = coinbaseStatement.Exec(winner.EntryHash, 0, Coinbase, addr, "", 0, "PEG", winner.Payout(), "")
	if err != nil {
		return err
	}

	if _, err = lookup.Exec(winner.EntryHash, 0, addr); err != nil {
		return err
	}

	return nil
}
