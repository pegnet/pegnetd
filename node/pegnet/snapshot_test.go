package pegnet_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"

	"github.com/stretchr/testify/assert"

	. "github.com/pegnet/pegnetd/node/pegnet"
)

func TestPegnet_SnapshotBalances(t *testing.T) {
	assert := assert.New(t)
	addresses := make([]factom.FAAddress, 10)
	for i := range addresses {
		copy(addresses[i][:], []byte{byte(i)})
	}

	checkBals := func(exps []uint64, bals []BalancesPair) error {
	checkLoop:
		for i, exp := range exps {
			add := addresses[i]
			for _, bal := range bals {
				if *bal.Address == add {
					if bal.Balances[fat2.PTickerPEG] != exp {
						return fmt.Errorf("add %x exp %d PEG, found %d", add[0], exp, bal.Balances[fat2.PTickerPEG])
					}
					continue checkLoop
				}
			}
			return fmt.Errorf("address %d not found", i)
		}
		return nil
	}

	// Open in memory sqlite
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Error(err)
	}
	var _ = db

	p := new(Pegnet)
	p.DB = db

	assert.NoError(p.CreateTableAddresses())

	addBalance := func(i int, amt int64) {
		// Add some balances
		tx, err := p.DB.Begin()
		assert.NoError(err)

		if amt > 0 {
			_, err = p.AddToBalance(tx, &addresses[i], fat2.PTickerPEG, uint64(amt))
			assert.NoError(err)
		}
		if amt < 0 {
			amt = amt * -1
			_, _, err = p.SubFromBalance(tx, &addresses[i], fat2.PTickerPEG, uint64(amt))
			assert.NoError(err)
		}
		_ = tx.Commit()
	}

	assert.NoError(p.SnapshotCurrent(p.DB))
	t.Run("no balances", func(t *testing.T) {
		// should be no balances since there is no snapshots currently

		bals, err := p.SelectSnapshotBalances(p.DB)
		assert.NoError(err)
		assert.Equal(0, len(bals))
	})

	// Balances all 0
	addBalance(0, 1e8)
	assert.NoError(p.SnapshotCurrent(p.DB))
	t.Run("no balances since past snapshot is empty", func(t *testing.T) {
		// Past snapshot is empty, current snapshot has 1.
		// Since the balance is the minimum of the sets, the balances should be 0
		bals, err := p.SelectSnapshotBalances(p.DB)
		assert.NoError(err)
		assert.Equal(0, len(bals))
	})

	// Past has Add[0] at 1e8 PEG
	assert.NoError(p.SnapshotCurrent(p.DB))

	t.Run("Add[0] balance should be 1 PEG (p&c match)", func(t *testing.T) {
		// Both snapshots have 1 peg in address[0]
		bals, err := p.SelectSnapshotBalances(p.DB)
		assert.NoError(err)
		assert.Equal(1, len(bals))
		if *bals[0].Address != addresses[0] {
			t.Errorf("expected address %x, got %x", addresses[0], bals[0].Address[0])
		}
		if bals[0].Balances[fat2.PTickerPEG] != 1e8 {
			t.Errorf("expected bal of 1e8, got %d", bals[0].Balances[fat2.PTickerPEG])
		}
	})

	// Past and Current Add[0] at 1e8 PEG
	addBalance(0, 1e8)
	addBalance(1, 5e8)
	assert.NoError(p.SnapshotCurrent(p.DB))
	t.Run("Add[0] balance should be 1 PEG (p&c mismatch)", func(t *testing.T) {
		// Both snapshots have 1 peg in address[0]
		bals, err := p.SelectSnapshotBalances(p.DB)
		assert.NoError(err)
		assert.Equal(1, len(bals))
		if *bals[0].Address != addresses[0] {
			t.Errorf("expected address %x, got %x", addresses[0], bals[0].Address[0])
		}
		if bals[0].Balances[fat2.PTickerPEG] != 1e8 {
			t.Errorf("expected bal of 1e8, got %d", bals[0].Balances[fat2.PTickerPEG])
		}
	})

	// Past Add[0] at 1e8 PEG
	// Current Add[0] at 2e8 PEG
	// Current Add[1] at 5e8 PEG

	addBalance(0, -2e8)
	assert.NoError(p.SnapshotCurrent(p.DB))
	t.Run("Add[0] at 5 PEG, Add[1] at 0 PEG (p&c mismatch)", func(t *testing.T) {
		// Both snapshots have 2 peg in address[0]
		// Past Add[1] has 5 PEG
		// Current Add[1] has 0 PEG
		bals, err := p.SelectSnapshotBalances(p.DB)
		assert.NoError(err)
		assert.Equal(2, len(bals))

		assert.NoError(checkBals([]uint64{0, 5e8}, bals))
	})

	// Past & Cur Add[0] at 0 PEG
	// Past & Cur Add[1] at 5e8 PEG

	addBalance(1, -2e8)
	assert.NoError(p.SnapshotCurrent(p.DB))
	t.Run("Add[0] at 3 PEG (p&c mismatch), Add[1] at 0 PEG", func(t *testing.T) {
		// Both snapshots have 2 peg in address[0]
		// Past Add[1] has 5 PEG
		// Current Add[1] has 0 PEG
		bals, err := p.SelectSnapshotBalances(p.DB)
		assert.NoError(err)
		assert.Equal(2, len(bals))

		assert.NoError(checkBals([]uint64{0, 3e8}, bals))
	})

}
