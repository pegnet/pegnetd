package pegnet_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/Factom-Asset-Tokens/factom"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pegnet/pegnetd/fat/fat2"
	. "github.com/pegnet/pegnetd/node/pegnet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPegnet() (*Pegnet, error) {
	// Taken from a combination of Pegnet.New() and Pegnet.Init()
	// just avoiding the config for now
	p := new(Pegnet)
	path := "/tmp/pegnet-tmp.db"
	_ = os.Remove(path)
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	p.DB = db
	err = p.CreateTableAddresses()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func tearDownPegnet(p *Pegnet) {
	_ = p.DB.Close()
	_ = os.Remove("/tmp/pegnet-tmp.db")
}

func TestPegnet_SelectBalance_Empty(t *testing.T) {
	p, err := setupPegnet()
	require.NoError(t, err)
	defer tearDownPegnet(p)

	var adr factom.FAAddress
	balance, err := p.SelectBalance(&adr, fat2.PTickerPEG)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), balance)
}

func TestPegnet_SelectBalance_InvalidTicker(t *testing.T) {
	p, err := setupPegnet()
	require.NoError(t, err)
	defer tearDownPegnet(p)

	var adr factom.FAAddress
	_, err = p.SelectBalance(&adr, fat2.PTicker(-1))
	assert.EqualError(t, err, "invalid token type")
	_, err = p.SelectBalance(&adr, fat2.PTickerInvalid)
	assert.EqualError(t, err, "invalid token type")
	_, err = p.SelectBalance(&adr, fat2.PTickerMax)
	assert.EqualError(t, err, "invalid token type")
	_, err = p.SelectBalance(&adr, fat2.PTicker(1000000000))
	assert.EqualError(t, err, "invalid token type")
}

func TestPegnet_SelectBalances_Empty(t *testing.T) {
	p, err := setupPegnet()
	require.NoError(t, err)
	defer tearDownPegnet(p)

	var adr factom.FAAddress
	balances, err := p.SelectBalances(&adr)
	require.NoError(t, err)
	require.Equal(t, int(fat2.PTickerMax)-1, len(balances), "Unexpected number of balances returned")
	for i := fat2.PTickerInvalid + 1; i < fat2.PTickerMax; i++ {
		assert.Equal(t, uint64(0), balances[i])
	}
}

func TestPegnet_AddToBalance(t *testing.T) {
	p, err := setupPegnet()
	require.NoError(t, err)
	defer tearDownPegnet(p)

	tx, err := p.DB.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	var adr factom.FAAddress
	_, err = p.AddToBalance(tx, &adr, fat2.PTickerPEG, 100)
	require.NoError(t, err)

	balance, err := p.SelectPendingBalance(tx, &adr, fat2.PTickerPEG)
	require.NoError(t, err)
	assert.Equal(t, uint64(100), balance, "Incorrect pending balance before tx.Commit()")

	balance, err = p.SelectBalance(&adr, fat2.PTickerPEG)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), balance, "Incorrect finalized balance before tx.Commit()")

	err = tx.Commit()
	require.NoError(t, err)

	balance, err = p.SelectBalance(&adr, fat2.PTickerPEG)
	require.NoError(t, err)
	assert.Equal(t, uint64(100), balance, "Incorrect finalized balance after tx.Commit()")
}

func TestPegnet_SubFromBalance(t *testing.T) {
	p, err := setupPegnet()
	require.NoError(t, err)
	defer tearDownPegnet(p)

	tx, err := p.DB.BeginTx(context.Background(), nil)
	require.NoError(t, err)

	var adr factom.FAAddress
	_, txErr, err := p.SubFromBalance(tx, &adr, fat2.PTickerPEG, 100)
	require.NoError(t, err)
	assert.EqualError(t, txErr, InsufficientBalanceErr.Error())

	_, err = p.AddToBalance(tx, &adr, fat2.PTickerPEG, 100)
	require.NoError(t, err)
	_, txErr, err = p.SubFromBalance(tx, &adr, fat2.PTickerPEG, 50)
	require.NoError(t, err)
	require.NoError(t, txErr)

	balance, err := p.SelectPendingBalance(tx, &adr, fat2.PTickerPEG)
	require.NoError(t, err)
	assert.Equal(t, uint64(50), balance, "Incorrect pending balance before tx.Commit()")

	balance, err = p.SelectBalance(&adr, fat2.PTickerPEG)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), balance, "Incorrect finalized balance before tx.Commit()")

	err = tx.Commit()
	require.NoError(t, err)

	balance, err = p.SelectBalance(&adr, fat2.PTickerPEG)
	require.NoError(t, err)
	assert.Equal(t, uint64(50), balance, "Incorrect finalized balance after tx.Commit()")
}
