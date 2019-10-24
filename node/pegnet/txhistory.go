package pegnet

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
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
	"blockorder"	INTEGER NOT NULL,
	"timestamp"		INTEGER NOT NULL,
	"executed"		INTEGER NOT NULL, -- -1 if failed, 0 if pending, height it was applied at otherwise

	PRIMARY KEY("entry_hash", "height"),
	FOREIGN KEY("height") REFERENCES "pn_grade"
);
CREATE INDEX IF NOT EXISTS "idx_transaction_history_entry_hash" ON "pn_transaction_history"("entry_hash");
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

type TxAction struct {
	Hash      *factom.Bytes32 `json:"hash"`
	Height    int64           `json:"height"`
	Timestamp time.Time       `json:"timestamp"`
	Executed  int32           `json:"executed"`
	TxIndex   int             `json:"txindex"`
	TxAction  HistoryAction   `json:"txaction"`

	ActionIndex int               `json:"actionindex"`
	FromAddress *factom.FAAddress `json:"fromaddress"`
	FromAsset   string            `json:"fromasset"`
	FromAmount  int64             `json:"fromamount"`
	ToAddress   *factom.FAAddress `json:"toaddress"`
	ToAsset     string            `json:"toasset"`
	ToAmount    int64             `json:"toamount"`
}

func _turnRowsIntoTxActions(rows *sql.Rows) ([]TxAction, error) {
	var actions []TxAction
	for rows.Next() {
		var a TxAction
		var ts int64
		var hash, from, to []byte
		err := rows.Scan(
			&hash, &a.Height, &ts, &a.Executed, // history
			&a.TxIndex, &a.TxAction, &from, &a.FromAsset, &a.FromAmount, // action
			&a.ActionIndex, &to, &a.ToAsset, &a.ToAmount) // data
		if err != nil {
			return nil, err
		}
		a.Hash = factom.NewBytes32(hash)
		a.Timestamp = time.Unix(ts, 0)
		var addr1, addr2 factom.FAAddress
		addr1 = factom.FAAddress(*factom.NewBytes32(from))
		a.FromAddress = &addr1
		addr2 = factom.FAAddress(*factom.NewBytes32(to))
		a.ToAddress = &addr2
		actions = append(actions, a)
	}
	return actions, nil
}

const baseTxCountQuery = `SELECT COUNT(*) FROM pn_transaction_history history, pn_transaction_history_action action, pn_transaction_history_action_data data
WHERE history.entry_hash = action.entry_hash AND action.entry_hash = data.entry_hash AND action.tx_index = data.tx_index AND (%s)`

const baseTxActionQuery = `SELECT 
	history.entry_hash, history.height, history.timestamp, history.executed,
	action.tx_index, action.action_type, action.address, action.asset, action.amount,
	data.tx_index_index, data.to_address, data.to_asset, data.to_amount
FROM pn_transaction_history history, pn_transaction_history_action action, pn_transaction_history_action_data data
WHERE history.entry_hash = action.entry_hash AND action.entry_hash = data.entry_hash AND action.tx_index = data.tx_index
	AND (%s)
ORDER BY %s
LIMIT 50 OFFSET %d`

func (p *Pegnet) SelectTransactionHistoryActionsByHash(hash *factom.Bytes32, offset int, descending bool) ([]TxAction, int, error) {
	var count int
	err := p.DB.QueryRow(fmt.Sprintf(baseTxCountQuery, "history.entry_hash = ?"), hash[:]).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return nil, 0, nil
	}

	if offset > count {
		return nil, 0, fmt.Errorf("offset too big")
	}

	order := "history.height, history.blockorder"
	if descending {
		order = "history.height DESC, history.blockorder DESC"
	}

	q := fmt.Sprintf(baseTxActionQuery, "history.entry_hash = ?", order, offset)

	rows, err := p.DB.Query(q, hash[:])
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	a, e := _turnRowsIntoTxActions(rows)
	return a, count, e
}

func (p *Pegnet) SelectTransactionHistoryActionsByAddress(addr *factom.FAAddress, offset int, descending bool) ([]TxAction, int, error) {
	var count int
	err := p.DB.QueryRow(fmt.Sprintf(baseTxCountQuery, "action.address = ? OR data.to_address = ?"), addr[:], addr[:]).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return nil, 0, nil
	}

	if offset > count {
		return nil, 0, fmt.Errorf("offset too big")
	}

	order := "history.height, history.blockorder"
	if descending {
		order = "history.height DESC, history.blockorder DESC"
	}

	q := fmt.Sprintf(baseTxActionQuery, "action.address = ? OR data.to_address = ?", order, offset)

	rows, err := p.DB.Query(q, addr[:], addr[:])
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	a, e := _turnRowsIntoTxActions(rows)
	return a, count, e
}

func (p *Pegnet) SelectTransactionHistoryActionsByHeight(height uint32, offset int, descending bool) ([]TxAction, int, error) {
	var count int
	err := p.DB.QueryRow(fmt.Sprintf(baseTxCountQuery, "history.height = ?"), height).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	if count == 0 {
		return nil, 0, nil
	}

	if offset > count {
		return nil, 0, fmt.Errorf("offset too big")
	}

	order := "history.height, history.blockorder"
	if descending {
		order = "history.height DESC, history.blockorder DESC"
	}

	q := fmt.Sprintf(baseTxActionQuery, "history.height = ?", order, offset)

	rows, err := p.DB.Query(q, height)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	a, e := _turnRowsIntoTxActions(rows)
	return a, count, e
}

func (p *Pegnet) SelectTransactionHistoryStatus(hash *factom.Bytes32) (uint32, uint32, error) {
	var height, executed uint32
	err := p.DB.QueryRow("SELECT height, executed FROM pn_transaction_history WHERE entry_hash = ?", hash[:]).Scan(&height, &executed)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return height, executed, nil
}

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

func (p *Pegnet) SetTransactionHistoryConvertedAmount(tx *sql.Tx, txbatch *fat2.TransactionBatch, index int, amount int64) error {
	stmt, err := tx.Prepare(`UPDATE "pn_transaction_history_action_data" SET to_amount = ? WHERE entry_hash = ? AND tx_index = ? AND tx_index_index = 0`)
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
	stmt, err := tx.Prepare(`INSERT INTO "pn_transaction_history"
                (entry_hash, height, timestamp, executed, blockorder) VALUES
                (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(txbatch.Entry.Hash[:], height, txbatch.Entry.Timestamp.Unix(), 0, blockorder)
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
			_, err = actionDataStatement.Exec(txbatch.Entry.Hash[:], index, 0, action.Input.Address[:], action.Conversion.String(), 0)
			if err != nil {
				return err
			}
		} else {
			for indexindex, data := range action.Transfers {
				_, err = actionDataStatement.Exec(txbatch.Entry.Hash[:], index, indexindex, data.Address[:], action.Input.Type.String(), data.Amount)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
