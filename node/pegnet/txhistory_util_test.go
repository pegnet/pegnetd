package pegnet

import (
	"reflect"
	"testing"
)

func TestHistoryQueryBuilder(t *testing.T) {
	type args struct {
		field   string
		options HistoryQueryOptions
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{ // only a single typed arg suffices since result of types is tested separately below
		{"empty", args{"", HistoryQueryOptions{}}, "", "", true},
		{"wrong field", args{"bad", HistoryQueryOptions{}}, "", "", true},
		{"entry hash, default args", args{"entry_hash", HistoryQueryOptions{}}, "SELECT COUNT(*) FROM pn_history_txbatch batch, pn_history_transaction tx WHERE batch.entry_hash = tx.entry_hash AND batch.entry_hash = ?", "SELECT batch.history_id, batch.entry_hash, batch.height, batch.timestamp, batch.executed,tx.tx_index, tx.action_type, tx.from_address, tx.from_asset, tx.from_amount, tx.outputs,tx.to_asset, tx.to_amount FROM pn_history_txbatch batch, pn_history_transaction tx WHERE batch.entry_hash = tx.entry_hash AND batch.entry_hash = ? ORDER BY batch.history_id ASC LIMIT 50 OFFSET 0", false},
		{"entry hash, offset", args{"entry_hash", HistoryQueryOptions{Offset: 123}}, "SELECT COUNT(*) FROM pn_history_txbatch batch, pn_history_transaction tx WHERE batch.entry_hash = tx.entry_hash AND batch.entry_hash = ?", "SELECT batch.history_id, batch.entry_hash, batch.height, batch.timestamp, batch.executed,tx.tx_index, tx.action_type, tx.from_address, tx.from_asset, tx.from_amount, tx.outputs,tx.to_asset, tx.to_amount FROM pn_history_txbatch batch, pn_history_transaction tx WHERE batch.entry_hash = tx.entry_hash AND batch.entry_hash = ? ORDER BY batch.history_id ASC LIMIT 50 OFFSET 123", false},
		{"entry hash, descending", args{"entry_hash", HistoryQueryOptions{Desc: true}}, "SELECT COUNT(*) FROM pn_history_txbatch batch, pn_history_transaction tx WHERE batch.entry_hash = tx.entry_hash AND batch.entry_hash = ?", "SELECT batch.history_id, batch.entry_hash, batch.height, batch.timestamp, batch.executed,tx.tx_index, tx.action_type, tx.from_address, tx.from_asset, tx.from_amount, tx.outputs,tx.to_asset, tx.to_amount FROM pn_history_txbatch batch, pn_history_transaction tx WHERE batch.entry_hash = tx.entry_hash AND batch.entry_hash = ? ORDER BY batch.history_id DESC LIMIT 50 OFFSET 0", false},
		{"entry hash, typed", args{"entry_hash", HistoryQueryOptions{FCTBurn: true, Coinbase: true}}, "SELECT COUNT(*) FROM pn_history_txbatch batch, pn_history_transaction tx WHERE (batch.entry_hash = tx.entry_hash AND batch.entry_hash = ?) AND tx.action_type IN(3,4)", "SELECT batch.history_id, batch.entry_hash, batch.height, batch.timestamp, batch.executed,tx.tx_index, tx.action_type, tx.from_address, tx.from_asset, tx.from_amount, tx.outputs,tx.to_asset, tx.to_amount FROM pn_history_txbatch batch, pn_history_transaction tx WHERE (batch.entry_hash = tx.entry_hash AND batch.entry_hash = ?) AND tx.action_type IN(3,4) ORDER BY batch.history_id ASC LIMIT 50 OFFSET 0", false},
		{"height, default args", args{"height", HistoryQueryOptions{}}, "SELECT COUNT(*) FROM pn_history_txbatch batch, pn_history_transaction tx WHERE batch.entry_hash = tx.entry_hash AND batch.height = ?", "SELECT batch.history_id, batch.entry_hash, batch.height, batch.timestamp, batch.executed,tx.tx_index, tx.action_type, tx.from_address, tx.from_asset, tx.from_amount, tx.outputs,tx.to_asset, tx.to_amount FROM pn_history_txbatch batch, pn_history_transaction tx WHERE batch.entry_hash = tx.entry_hash AND batch.height = ? ORDER BY batch.history_id ASC LIMIT 50 OFFSET 0", false},
		{"address, default args", args{"address", HistoryQueryOptions{}}, "SELECT COUNT(*) FROM pn_history_lookup WHERE address = ?", "SELECT batch.history_id, batch.entry_hash, batch.height, batch.timestamp, batch.executed,tx.tx_index, tx.action_type, tx.from_address, tx.from_asset, tx.from_amount, tx.outputs,tx.to_asset, tx.to_amount FROM pn_history_lookup lookup, pn_history_txbatch batch, pn_history_transaction tx WHERE lookup.address = ? AND lookup.entry_hash = tx.entry_hash AND lookup.tx_index = tx.tx_index AND batch.entry_hash = tx.entry_hash ORDER BY batch.history_id ASC LIMIT 50 OFFSET 0", false},
		{"address, typed", args{"address", HistoryQueryOptions{Conversion: true, Transfer: true}}, "SELECT COUNT(*) FROM pn_history_lookup lookup, pn_history_transaction tx WHERE (lookup.address = ? AND lookup.entry_hash = tx.entry_hash AND lookup.tx_index = tx.tx_index) AND tx.action_type IN(1,2)", "SELECT batch.history_id, batch.entry_hash, batch.height, batch.timestamp, batch.executed,tx.tx_index, tx.action_type, tx.from_address, tx.from_asset, tx.from_amount, tx.outputs,tx.to_asset, tx.to_amount FROM pn_history_lookup lookup, pn_history_txbatch batch, pn_history_transaction tx WHERE (lookup.address = ? AND lookup.entry_hash = tx.entry_hash AND lookup.tx_index = tx.tx_index AND batch.entry_hash = tx.entry_hash) AND tx.action_type IN(1,2) ORDER BY batch.history_id ASC LIMIT 50 OFFSET 0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := historyQueryBuilder(tt.args.field, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("HistoryQueryBuilder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HistoryQueryBuilder() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("HistoryQueryBuilder() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_historyActionPicker(t *testing.T) {
	type args struct {
		tx   bool
		conv bool
		coin bool
		burn bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"set-0", args{false, false, false, false}, nil},
		{"set-1", args{false, false, false, true}, []string{"4"}},
		{"set-2", args{false, false, true, false}, []string{"3"}},
		{"set-3", args{false, false, true, true}, []string{"3", "4"}},
		{"set-4", args{false, true, false, false}, []string{"2"}},
		{"set-5", args{false, true, false, true}, []string{"2", "4"}},
		{"set-6", args{false, true, true, false}, []string{"2", "3"}},
		{"set-7", args{false, true, true, true}, []string{"2", "3", "4"}},
		{"set-8", args{true, false, false, false}, []string{"1"}},
		{"set-9", args{true, false, false, true}, []string{"1", "4"}},
		{"set-10", args{true, false, true, false}, []string{"1", "3"}},
		{"set-11", args{true, false, true, true}, []string{"1", "3", "4"}},
		{"set-12", args{true, true, false, false}, []string{"1", "2"}},
		{"set-13", args{true, true, false, true}, []string{"1", "2", "4"}},
		{"set-14", args{true, true, true, false}, []string{"1", "2", "3"}},
		{"set-15", args{true, true, true, true}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := historyActionPicker(tt.args.tx, tt.args.conv, tt.args.coin, tt.args.burn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("historyActionPicker() = %v, want %v", got, tt.want)
			}
		})
	}
}
