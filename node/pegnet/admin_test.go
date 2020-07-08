package pegnet_test

import (
	"database/sql"
	"testing"

	"github.com/pegnet/pegnetd/node/pegnet"
)

func TestPegnet_CheckHardForks(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Error(err)
	}

	p := new(pegnet.Pegnet)
	p.DB = db
	if err := p.CreateTableSyncVersion(); err != nil {
		t.Error(err)
	}

	t.Run("test blank db", func(t *testing.T) {
		if err := p.CheckHardForks(p.DB); err != nil {
			t.Errorf("blank db error: %v", err)
		}
	})

	pegnet.Hardforks = []pegnet.ForkEvent{
		{ActivationHeight: 0, MinimumVersion: 0},
		{ActivationHeight: 1, MinimumVersion: 1},
		{ActivationHeight: 2, MinimumVersion: 2},
		{ActivationHeight: 3, MinimumVersion: 3},
		{ActivationHeight: 4, MinimumVersion: 4},
		{ActivationHeight: 5, MinimumVersion: 5},
		{ActivationHeight: 9, MinimumVersion: 9},
		{ActivationHeight: 15, MinimumVersion: 15},
		{ActivationHeight: 20, MinimumVersion: 20},
		{ActivationHeight: 1000, MinimumVersion: 1000},
	}

	// Height 0-10
	t.Run("test all versions updated", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			pegnet.PegnetdSyncVersion = i
			_ = p.MarkHeightSynced(p.DB, uint32(i))
		}

		if err := p.CheckHardForks(p.DB); err != nil {
			t.Errorf("correct db error: %v", err)
		}
	})

	// Height 15
	t.Run("test version high above", func(t *testing.T) {
		pegnet.PegnetdSyncVersion = 500
		_ = p.MarkHeightSynced(p.DB, uint32(15))

		if err := p.CheckHardForks(p.DB); err != nil {
			t.Errorf("correct db error: %v", err)
		}
	})

	// Height 20
	t.Run("test version 1 behind", func(t *testing.T) {
		pegnet.PegnetdSyncVersion = 19
		_ = p.MarkHeightSynced(p.DB, uint32(20))

		if err := p.CheckHardForks(p.DB); err == nil {
			t.Errorf("expected error, found none")
		}
	})
}
