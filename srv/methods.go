// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package srv

import (
	"bytes"
	"encoding/json"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/Factom-Asset-Tokens/fatd/db/entries"
	"github.com/Factom-Asset-Tokens/fatd/engine"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/pegnet/pegnetd/fat/fat2"
)

var c = flag.FactomClient

var jrpcMethods = jrpc.MethodMap{
	"get-transaction":       getTransaction(false),
	"get-transaction-entry": getTransaction(true),
	"get-pegnet-balances":   getPegnetBalances,

	"send-transaction": sendTransaction,

	"get-sync-status": getSyncStatus,
}

type ResultGetTransaction struct {
	Hash      *factom.Bytes32 `json:"entryhash"`
	Timestamp int64           `json:"timestamp"`
	TxIndex   uint64          `json:"txindex,omitempty"`
	Tx        interface{}     `json:"data"`
}

func getTransaction(getEntry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		params := ParamsGetTransaction{}
		chain, put, err := validate(data, &params)
		if err != nil {
			return err
		}
		defer put()

		// TODO: use pegnet specific database code here to fetch an entry
		entry, err := entries.SelectValidByHash(chain.Conn, params.Hash)
		if err != nil {
			panic(err)
		}
		if !entry.IsPopulated() {
			return ErrorTransactionNotFound
		}

		if getEntry {
			return entry
		}

		result := ResultGetTransaction{
			Hash:      entry.Hash,
			Timestamp: entry.Timestamp.Unix(),
		}

		tx := fat2.NewTransactionBatch(entry)
		if err := tx.UnmarshalEntry(); err != nil {
			panic(err)
		}
		result.Tx = tx
		return result
	}
}

type ResultGetPegnetBalances map[fat2.PTicker]uint64

func (r ResultGetPegnetBalances) MarshalJSON() ([]byte, error) {
	strMap := make(map[string]uint64, len(r))
	for ticker, balance := range r {
		strMap[ticker.String()] = balance
	}
	return json.Marshal(strMap)
}
func (r *ResultGetPegnetBalances) UnmarshalJSON(data []byte) error {
	var strMap map[string]uint64
	if err := json.Unmarshal(data, &strMap); err != nil {
		return err
	}
	*r = make(map[fat2.PTicker]uint64, len(strMap))
	for str, balance := range strMap {
		var ticker fat2.PTicker
		if err := fat2.PTicker.UnmarshalJSON(ticker, []byte(str)); err != nil {
			return err
		}
		(*r)[ticker] = balance
	}
	return nil
}

func getPegnetBalances(data json.RawMessage) interface{} {
	params := ParamsGetPegnetBalances{}
	if _, _, err := validate(data, &params); err != nil {
		return err
	}
	balances := make(ResultGetPegnetBalances, int(fat2.PTickerMax))
	for i := 1; i < int(fat2.PTickerMax); i++ {
		//_, balance, err := db.SelectBalances(params.Address)
		//if err != nil {
		//	panic(err)
		//}
		//if balance > 0 {
		//	balances[fat2.PTicker(i)] = balance
		//}
	}
	return balances
}

func sendTransaction(data json.RawMessage) interface{} {
	params := ParamsSendTransaction{}
	chain, put, err := validate(data, &params)
	if err != nil {
		return err
	}
	defer put()

	// TODO: srv.sendTransaction(): use EC address from a config file
	if !params.DryRun && factom.Bytes32(flag.EsAdr).IsZero() {
		return ErrorNoEC
	}

	entry := params.Entry()
	txErr, err := attemptApplyFAT2TxBatch(chain, entry)
	if err != nil {
		panic(err)
	}
	if txErr != nil {
		err := ErrorInvalidTransaction
		err.Data = txErr.Error()
		return err
	}

	var txID *factom.Bytes32
	if !params.DryRun {
		balance, err := flag.ECAdr.GetBalance(c)
		if err != nil {
			panic(err)
		}
		cost, err := entry.Cost()
		if err != nil {
			rerr := ErrorInvalidTransaction
			rerr.Data = err.Error()
			return rerr
		}
		if balance < uint64(cost) {
			return ErrorNoEC
		}
		txID, err = entry.ComposeCreate(c, flag.EsAdr)
		if err != nil {
			panic(err)
		}
	}

	return struct {
		ChainID *factom.Bytes32 `json:"chainid"`
		TxID    *factom.Bytes32 `json:"txid,omitempty"`
		Hash    *factom.Bytes32 `json:"entryhash"`
	}{ChainID: chain.ID, TxID: txID, Hash: entry.Hash}
}
func attemptApplyFAT2TxBatch(chain *engine.Chain, e factom.Entry) (txErr, err error) {
	txBatch := fat2.NewTransactionBatch(e)
	if txErr = txBatch.Validate(); txErr != nil {
		return
	}
	// TODO: check this entry never been put in chain before
	//valid, err := entries.CheckUniquelyValid(chain.Conn, 0, e.Hash)
	//if err != nil {
	//	return
	//}
	//if !valid {
	//	txErr = fmt.Errorf("replay: hash previously marked valid")
	//	return
	//}

	// TODO: Check all input balances

	return
}

type ResultGetSyncStatus struct {
	Sync    uint32 `json:"syncheight"`
	Current uint32 `json:"factomheight"`
}

func getSyncStatus(data json.RawMessage) interface{} {
	sync, current := engine.GetSyncStatus()
	return ResultGetSyncStatus{Sync: sync, Current: current}
}

func validate(data json.RawMessage, params Params) (*engine.Chain, func(), error) {
	if params == nil {
		if len(data) > 0 {
			return nil, nil, jrpc.InvalidParams(`no "params" accepted`)
		}
		return nil, nil, nil
	}
	if len(data) == 0 {
		return nil, nil, params.IsValid()
	}
	if err := unmarshalStrict(data, params); err != nil {
		return nil, nil, jrpc.InvalidParams(err.Error())
	}
	if err := params.IsValid(); err != nil {
		return nil, nil, err
	}
	if params.HasIncludePending() && flag.DisablePending {
		return nil, nil, ErrorPendingDisabled
	}
	chainID := params.ValidChainID()
	if chainID != nil {
		chain := engine.Chains.Get(chainID)
		if !chain.IsIssued() {
			return nil, nil, ErrorTokenNotFound
		}
		if params.HasIncludePending() {
			chain.ApplyPending()
		}
		conn, put := chain.Get()
		chain.Conn = conn
		return &chain, put, nil
	}
	return nil, nil, nil
}

func unmarshalStrict(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
