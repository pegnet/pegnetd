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

type HistoryAction int32

const (
	Invalid HistoryAction = iota
	Transfer
	Conversion
	Coinbase
	FCTBurn
)
const createTableTxHistoryBatch = `CREATE TABLE IF NOT EXISTS "pn_history_txbatch" (
	"history_id"	INTEGER PRIMARY KEY,
	"entry_hash"    BLOB NOT NULL,
	"height"        INTEGER NOT NULL, -- height the tx is in
	"blockorder"	INTEGER NOT NULL,
	"timestamp"		INTEGER NOT NULL,
	"executed"		INTEGER NOT NULL, -- -1 if failed, 0 if pending, height it was applied at otherwise

	UNIQUE("entry_hash", "height"),
	FOREIGN KEY("height") REFERENCES "pn_grade"
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

const insertLookupQuery = `INSERT INTO pn_history_lookup (entry_hash, tx_index, address) VALUES (?, ?, ?) ON CONFLICT DO NOTHING;`

type HistoryTransaction struct {
	Hash      *factom.Bytes32 `json:"hash"`
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
type HistoryTransactionOutput struct {
	Address *factom.FAAddress `json:"address"`
	Amount  int64             `json:"amount"`
}

func _turnRowsIntoHistoryTransactions(rows *sql.Rows) ([]HistoryTransaction, error) {
	var actions []HistoryTransaction
	for rows.Next() {
		var tx HistoryTransaction
		var ts, id int64
		var hash, from, outputs []byte
		err := rows.Scan(
			&id, &hash, &tx.Height, &ts, &tx.Executed, // history
			&tx.TxIndex, &tx.TxAction, &from, &tx.FromAsset, &tx.FromAmount, // action
			&outputs, &tx.ToAsset, &tx.ToAmount) // data
		if err != nil {
			return nil, err
		}
		tx.Hash = factom.NewBytes32(hash)
		tx.Timestamp = time.Unix(ts, 0)
		var addr factom.FAAddress
		addr = factom.FAAddress(*factom.NewBytes32(from))
		tx.FromAddress = &addr

		if tx.TxAction == Transfer {
			var output []HistoryTransactionOutput
			if err = json.Unmarshal(outputs, &output); err != nil { // should never fail unless database data is corrupt
				return nil, fmt.Errorf("database corruption %d %v", id, err)
			}
			tx.Outputs = output
		}

		actions = append(actions, tx)
	}
	return actions, nil
}

const baseTxCountQuery = `SELECT COUNT(*) FROM pn_history_txbatch batch, pn_history_transaction tx
WHERE batch.entry_hash = tx.entry_hash AND (%s)`

const baseTxActionQuery = `SELECT 
	batch.history_id, batch.entry_hash, batch.height, batch.timestamp, batch.executed,
	tx.tx_index, tx.action_type, tx.from_address, tx.from_asset, tx.from_amount, tx.outputs,
	tx.to_asset, tx.to_amount
FROM pn_history_txbatch batch, pn_history_transaction tx
WHERE batch.entry_hash = tx.entry_hash AND (%s)
ORDER BY batch.history_id %s
LIMIT 50 OFFSET %d`

func (p *Pegnet) SelectTransactionHistoryActionsByHash(hash *factom.Bytes32, offset int, descending bool) ([]HistoryTransaction, int, error) {
	var count int
	err := p.DB.QueryRow(fmt.Sprintf(baseTxCountQuery, "batch.entry_hash = ?"), hash[:]).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return nil, 0, nil
	}

	if offset > count {
		return nil, 0, fmt.Errorf("offset too big")
	}

	order := "ASC"
	if descending {
		order = "DESC"
	}

	q := fmt.Sprintf(baseTxActionQuery, "batch.entry_hash = ?", order, offset)

	rows, err := p.DB.Query(q, hash[:])
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	a, e := _turnRowsIntoHistoryTransactions(rows)
	return a, count, e
}

const addressQuery = `SELECT 
	batch.history_id, batch.entry_hash, batch.height, batch.timestamp, batch.executed,
	tx.tx_index, tx.action_type, tx.from_address, tx.from_asset, tx.from_amount, tx.outputs,
	tx.to_asset, tx.to_amount
FROM pn_history_lookup lookup, pn_history_txbatch batch, pn_history_transaction tx
WHERE lookup.address = ? AND lookup.entry_hash = tx.entry_hash AND lookup.tx_index = tx.tx_index AND batch.entry_hash = tx.entry_hash
ORDER BY batch.history_id %s
LIMIT 50 OFFSET %d`
const addressQueryCount = `SELECT COUNT(*) FROM pn_history_lookup lookup WHERE lookup.address = ?`

func (p *Pegnet) SelectTransactionHistoryActionsByAddress(addr *factom.FAAddress, offset int, descending bool) ([]HistoryTransaction, int, error) {
	var count int
	err := p.DB.QueryRow(addressQueryCount, addr[:]).Scan(&count)
	if err != nil {
		return nil, 0, err
	}
	if count == 0 {
		return nil, 0, nil
	}
	if offset > count {
		return nil, 0, fmt.Errorf("offset too big")
	}

	order := "ASC"
	if descending {
		order = "DESC"
	}
	q := fmt.Sprintf(addressQuery, order, offset)

	rows, err := p.DB.Query(q, addr[:])
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	a, e := _turnRowsIntoHistoryTransactions(rows)
	return a, count, e
}

func (p *Pegnet) SelectTransactionHistoryActionsByHeight(height uint32, offset int, descending bool) ([]HistoryTransaction, int, error) {
	var count int
	err := p.DB.QueryRow(fmt.Sprintf(baseTxCountQuery, "batch.height = ?"), height).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return nil, 0, nil
	}

	if offset > count {
		return nil, 0, fmt.Errorf("offset too big")
	}

	order := "ASC"
	if descending {
		order = "DESC"
	}

	q := fmt.Sprintf(baseTxActionQuery, "batch.height = ?", order, offset)

	rows, err := p.DB.Query(q, height)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	a, e := _turnRowsIntoHistoryTransactions(rows)
	return a, count, e
}

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
			outputs := make([]HistoryTransactionOutput, len(action.Transfers))
			for i, transfer := range action.Transfers {
				outputs[i] = HistoryTransactionOutput{Address: &transfer.Address, Amount: int64(transfer.Amount)}
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
