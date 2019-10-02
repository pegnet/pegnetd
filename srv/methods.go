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
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pegnet/pegnetd/node"

	"github.com/AdamSLevy/jsonrpc2"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/factom"
	"github.com/pegnet/pegnetd/fat/fat2"
)

func (s *APIServer) jrpcMethods() jrpc.MethodMap {
	return jrpc.MethodMap{
		"get-transaction":       s.getTransaction(false),
		"get-transaction-entry": s.getTransaction(true),
		"get-pegnet-balances":   s.getPegnetBalances,

		"send-transaction": sendTransaction,

		"get-sync-status": s.getSyncStatus,
	}

}

type ResultGetTransaction struct {
	Hash      *factom.Bytes32 `json:"entryhash"`
	Timestamp int64           `json:"timestamp"`
	TxIndex   uint64          `json:"txindex,omitempty"`
	Tx        interface{}     `json:"data"`
}

func (s *APIServer) getTransaction(getEntry bool) jrpc.MethodFunc {
	return func(data json.RawMessage) interface{} {
		//params := ParamsGetTransaction{}
		//chain, put, err := validate(data, &params)
		//if err != nil {
		//	return err
		//}
		//defer put()

		// TODO: use pegnet specific database code here to fetch an entry
		//entry, err := entries.SelectValidByHash(chain.Conn, params.Hash)
		//if err != nil {
		//	panic(err)
		//}
		//if !entry.IsPopulated() {
		//	return ErrorTransactionNotFound
		//}
		//
		//if getEntry {
		//	return entry
		//}
		//
		//result := ResultGetTransaction{
		//	Hash:      entry.Hash,
		//	Timestamp: entry.Timestamp.Unix(),
		//}
		//
		//tx := fat2.NewTransactionBatch(entry)
		//if err := tx.UnmarshalEntry(); err != nil {
		//	panic(err)
		//}
		//result.Tx = tx
		//return result
		return nil
	}
}

// TODO: This is incompatible with FAT.
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
		ticker := new(fat2.PTicker)
		if err := ticker.UnmarshalJSON([]byte(str)); err != nil {
			return err
		}
		//if err := fat2.PTicker.UnmarshalJSON(&ticker, []byte(str)); err != nil {
		//	return err
		//}
		(*r)[*ticker] = balance
	}
	return nil
}

func (s *APIServer) getPegnetBalances(data json.RawMessage) interface{} {
	params := ParamsGetPegnetBalances{}
	if _, _, err := validate(data, &params); err != nil {
		return err
	}
	bals, err := s.Node.Pegnet.SelectBalances(params.Address)
	if err == sql.ErrNoRows {
		return ErrorAddressNotFound
	}
	if err != nil {
		return jsonrpc2.InternalError
	}
	return ResultGetPegnetBalances(bals)
}

func sendTransaction(data json.RawMessage) interface{} {
	//params := ParamsSendTransaction{}
	//chain, put, err := validate(data, &params)
	//if err != nil {
	//	return err
	//}
	//defer put()
	//
	//// TODO: srv.sendTransaction(): use EC address from a config file
	//if !params.DryRun && factom.Bytes32(flag.EsAdr).IsZero() {
	//	return ErrorNoEC
	//}
	//
	//entry := params.Entry()
	//txErr, err := attemptApplyFAT2TxBatch(chain, entry)
	//if err != nil {
	//	panic(err)
	//}
	//if txErr != nil {
	//	err := ErrorInvalidTransaction
	//	err.Data = txErr.Error()
	//	return err
	//}
	//
	//var txID *factom.Bytes32
	//if !params.DryRun {
	//	balance, err := flag.ECAdr.GetBalance(c)
	//	if err != nil {
	//		panic(err)
	//	}
	//	cost, err := entry.Cost()
	//	if err != nil {
	//		rerr := ErrorInvalidTransaction
	//		rerr.Data = err.Error()
	//		return rerr
	//	}
	//	if balance < uint64(cost) {
	//		return ErrorNoEC
	//	}
	//	txID, err = entry.ComposeCreate(c, flag.EsAdr)
	//	if err != nil {
	//		panic(err)
	//	}
	//}

	//return struct {
	//	ChainID *factom.Bytes32 `json:"chainid"`
	//	TxID    *factom.Bytes32 `json:"txid,omitempty"`
	//	Hash    *factom.Bytes32 `json:"entryhash"`
	//}{ChainID: chain.ID, TxID: txID, Hash: entry.Hash}
	// TODO: Implement this, and probably allow the user to provide a self signed commit for a shared node
	return nil
}

//func attemptApplyFAT2TxBatch(chain *engine.Chain, e factom.Entry) (txErr, err error) {
//	txBatch := fat2.NewTransactionBatch(e)
//	if txErr = txBatch.Validate(); txErr != nil {
//		return
//	}
//	// TODO: check this entry never been put in chain before
//	//valid, err := entries.CheckUniquelyValid(chain.Conn, 0, e.Hash)
//	//if err != nil {
//	//	return
//	//}
//	//if !valid {
//	//	txErr = fmt.Errorf("replay: hash previously marked valid")
//	//	return
//	//}
//
//	// TODO: Check all input balances
//
//	return
//}

type ResultGetSyncStatus struct {
	Sync    uint32 `json:"syncheight"`
	Current int32  `json:"factomheight"`
}

func (s *APIServer) getSyncStatus(data json.RawMessage) interface{} {
	heights := new(factom.Heights)
	err := heights.Get(s.Node.FactomClient)
	if err != nil {
		return ResultGetSyncStatus{Sync: s.Node.GetCurrentSync(), Current: -1}
	}
	return ResultGetSyncStatus{Sync: s.Node.GetCurrentSync(), Current: int32(heights.DirectoryBlock)}
}

// TODO: Re-eval this function. The chain data that is supplied needs to be reimplemented
//		return was (*engine.Chain, func(), error)
func validate(data json.RawMessage, params Params) (interface{}, func(), error) {
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
	//if params.HasIncludePending() && flag.DisablePending {
	//	return nil, nil, ErrorPendingDisabled
	//}
	chainID := params.ValidChainID()
	if chainID != nil {
		fmt.Printf("%x\n", chainID)
		if *chainID != node.TransactionChain {
			return nil, nil, ErrorTokenNotFound
		}
		// TODO: Do we need to stub out any of the chain fields?
		//chain := engine.Chains.Get(chainID)
		//if !chain.IsIssued() {
		//	return nil, nil, ErrorTokenNotFound
		//}
		//if params.HasIncludePending() {
		//	chain.ApplyPending()
		//}
		//conn, put := chain.Get()
		//chain.Conn = conn
		//return &chain, put, nil
	}

	// If there is no chain, then we can't really validate it since we aren't fatd.
	// The chainid is just to be compatible, but in reality it means nothing to us.
	return nil, nil, nil
}

func unmarshalStrict(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	d := json.NewDecoder(b)
	d.DisallowUnknownFields()
	return d.Decode(v)
}
