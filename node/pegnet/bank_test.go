package pegnet_test

import (
	"database/sql"
	"testing"

	"github.com/pegnet/pegnetd/node/pegnet"
)

func TestPegnet_BankTable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Error(err)
	}

	p := new(pegnet.Pegnet)
	p.DB = db
	if err := p.CreateTableBank(); err != nil {
		t.Error(err)
	}

	t.Run("no row", func(t *testing.T) {
		entry, err := p.SelectBankEntry(nil, 10)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if entry.Height != -1 || entry.BankAmount != -1 {
			t.Errorf("expected all values to be -1")
		}
	})

	t.Run("duplicate entry", func(t *testing.T) {
		err := p.InsertBankAmount(nil, 10, 5000)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		err = p.InsertBankAmount(nil, 10, 5000)
		if err == nil {
			t.Errorf("expected error, got none")
		}
	})

	t.Run("update not exists entry", func(t *testing.T) {
		err := p.UpdateBankEntry(nil, 9, 0, 0)
		if err == nil {
			t.Errorf("expected error, got none")
		}

		err = p.InsertBankAmount(nil, 9, 5000)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		err = p.UpdateBankEntry(nil, 9, 0, 0)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
