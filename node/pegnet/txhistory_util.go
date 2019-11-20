package pegnet

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"
)

// HistoryAction are the different types of actions inside the history
type HistoryAction int32

const (
	// Invalid is used for debugging
	Invalid HistoryAction = iota
	// Transfer is a 1:n transfer of pegged assets from one address to another
	Transfer
	// Conversion is a conversion of pegged assets
	Conversion
	// Coinbase is a miner reward payout
	Coinbase
	// FCTBurn is a pFCT payout for burning FCT on factom
	FCTBurn
)

// QueryLimit is the amount of transactions to return in one query
const QueryLimit = 50

func historyActionPicker(tx, conv, coin, burn bool) []string {
	if tx == conv && conv == coin && coin == burn {
		return nil
	}

	var actions []string
	if tx {
		actions = append(actions, strconv.Itoa(int(Transfer)))
	}
	if conv {
		actions = append(actions, strconv.Itoa(int(Conversion)))
	}
	if coin {
		actions = append(actions, strconv.Itoa(int(Coinbase)))
	}
	if burn {
		actions = append(actions, strconv.Itoa(int(FCTBurn)))
	}

	return actions
}

// HistoryQueryOptions contains the data of what to query for the query builder
type HistoryQueryOptions struct {
	Offset     int
	Desc       bool
	Transfer   bool
	Conversion bool
	Coinbase   bool
	FCTBurn    bool
	Asset      string

	// UseTxIndex is set if specifying a specific tx in the batch.
	// Because 0 is a valid tx index, we want the uninitialized value
	// to be "off"
	UseTxIndex bool
	TxIndex    int
}

const historyQueryFields = "batch.history_id, batch.entry_hash, batch.height, batch.timestamp, batch.executed," +
	"tx.tx_index, tx.action_type, tx.from_address, tx.from_asset, tx.from_amount, tx.outputs," +
	"tx.to_asset, tx.to_amount"

// historyQueryBuilder generates a count and data query for the given options
func historyQueryBuilder(field string, options HistoryQueryOptions) (string, string, error) {
	order := "ORDER BY batch.history_id ASC"
	if options.Desc {
		order = "ORDER BY batch.history_id DESC"
	}

	limit := fmt.Sprintf("LIMIT %d OFFSET %d", QueryLimit, options.Offset)

	types := historyActionPicker(options.Transfer, options.Conversion, options.Coinbase, options.FCTBurn)

	var from, where, fromCount, whereCount string
	switch field {
	case "address":
		if types != nil || options.Asset != "" {
			fromCount = "pn_history_lookup lookup, pn_history_transaction tx"
			whereCount = "lookup.address = ? AND lookup.entry_hash = tx.entry_hash AND lookup.tx_index = tx.tx_index"
		} else {
			fromCount = "pn_history_lookup"
			whereCount = "address = ?"
		}
		from = "pn_history_lookup lookup, pn_history_txbatch batch, pn_history_transaction tx"
		where = "lookup.address = ? AND lookup.entry_hash = tx.entry_hash AND lookup.tx_index = tx.tx_index AND batch.entry_hash = tx.entry_hash"
	case "entry_hash":
		fallthrough
	case "height":
		from = "pn_history_txbatch batch, pn_history_transaction tx"
		where = fmt.Sprintf("batch.entry_hash = tx.entry_hash AND batch.%s = ?", field)
		fromCount = from
		whereCount = where
	default:
		return "", "", fmt.Errorf("developer error - unimplemented history query builder field")
	}

	// Only select the txindex. Only works with entry_hash field
	if options.UseTxIndex && field == "entry_hash" {
		where += fmt.Sprintf(" AND tx.tx_index = %d", options.TxIndex)
	}

	if options.Asset != "" {
		tick := fat2.StringToTicker(options.Asset)
		if tick.String() == "invalid token type" {
			return "", "", fmt.Errorf("invalid asset specified")
		}
		where += fmt.Sprintf(" AND (tx.from_asset = '%s' OR tx.to_asset = '%s')", options.Asset, options.Asset)
		whereCount += fmt.Sprintf(" AND (tx.from_asset = '%s' OR tx.to_asset = '%s')", options.Asset, options.Asset)
	}

	if types != nil {
		where = fmt.Sprintf("(%s) AND tx.action_type IN(%s)", where, strings.Join(types, ","))
		whereCount = fmt.Sprintf("(%s) AND tx.action_type IN(%s)", whereCount, strings.Join(types, ","))
	}

	return fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", fromCount, whereCount),
		fmt.Sprintf("SELECT %s FROM %s WHERE %s %s %s", historyQueryFields, from, where, order, limit), nil
}

// helper function for sql results of a query builder's data query
func turnRowsIntoHistoryTransactions(rows *sql.Rows) ([]HistoryTransaction, error) {
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
		tx.TxID = FormatTxID(tx.TxIndex, tx.Hash.String())
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
