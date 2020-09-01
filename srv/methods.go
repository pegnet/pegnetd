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
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"

	jrpc "github.com/AdamSLevy/jsonrpc2/v13"
	"github.com/Factom-Asset-Tokens/factom"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/pegnet/pegnet/modules/conversions"
	"github.com/pegnet/pegnetd/config"
	"github.com/pegnet/pegnetd/fat/fat2"
	"github.com/pegnet/pegnetd/node"
	"github.com/pegnet/pegnetd/node/pegnet"
)

func (s *APIServer) jrpcMethods() jrpc.MethodMap {
	return jrpc.MethodMap{
		"get-rich-list":          s.getRichList,
		"get-global-rich-list":   s.getGlobalRichList,
		"get-miner-distribution": s.getMiningDominance,
		"get-bank":               s.getBank,
		"get-transactions":       s.getTransactions(false),
		"get-transaction-status": s.getTransactionStatus,
		"get-transaction":        s.getTransactions(true),
		"get-pegnet-balances":    s.getPegnetBalances,
		"get-pegnet-issuance":    s.getPegnetIssuance,
		"send-transaction":       s.sendTransaction,

		"get-sync-status": s.getSyncStatus,
		"properties":      s.properties,

		"get-pegnet-rates": s.getRecentPegnetRates,
		"get-stats":        s.getStats,
	}

}

type PegnetdProperties struct {
	BuildVersion  string `json:"buildversion"`
	BuildCommit   string `json:"buildcommit"`
	SQLiteVersion string `json:"sqliteversion"`
	GolangVersion string `json:"golang"`
}

func (APIServer) properties(_ context.Context, data json.RawMessage) interface{} {
	sqliteVersion, _, _ := sqlite3.Version()
	return PegnetdProperties{
		BuildVersion:  config.CompiledInVersion,
		BuildCommit:   config.CompiledInBuild,
		SQLiteVersion: sqliteVersion,
		GolangVersion: runtime.Version(),
	}
}

func (s *APIServer) getBank(ctx context.Context, data json.RawMessage) interface{} {
	params := ParamsGetBank{}
	_, _, err := validate(data, &params)
	if err != nil {
		return err
	}

	if params.Height == 0 {
		synced, err := s.Node.Pegnet.SelectSynced(ctx, s.Node.Pegnet.DB)
		if err != nil {
			return err
		}
		params.Height = int32(synced.Synced)
	}

	if params.Height < int32(node.V4OPRUpdate) {
		return jrpc.ErrorInvalidParams(fmt.Sprintf("the height %d is below the activation height (%d) of this feature", params.Height, node.V4OPRUpdate))
	}
	result, err := s.Node.Pegnet.SelectBankEntry(nil, params.Height)
	if err != nil {
		return err
	}
	return result
}

// getMiningDominance returns the representation of rewarded miners for a given
// block range
func (s *APIServer) getMiningDominance(ctx context.Context, data json.RawMessage) interface{} {
	params := ParamsGetMiningDominance{}
	_, _, err := validate(data, &params)
	if err != nil {
		return err
	}

	if params.Start == 0 && params.Stop < 0 {
		// If the start is 0, and stop is negative, then the user is requesting
		// the last STOP blocks
		synced, err := s.Node.Pegnet.SelectSynced(ctx, s.Node.Pegnet.DB)
		if err != nil {
			return err
		}
		params.Start = int(synced.Synced) + params.Stop
		params.Stop = int(synced.Synced)
	} else if params.Stop == 0 {
		// If the stop is 0, then the stop is the end.
		synced, err := s.Node.Pegnet.SelectSynced(ctx, s.Node.Pegnet.DB)
		if err != nil {
			return err
		}
		params.Stop = int(synced.Synced)
	}

	result, err := s.Node.Pegnet.SelectMinerDominance(ctx, params.Start, params.Stop)
	if err != nil {
		return err
	}
	return result
}

type ResultGlobalRichList struct {
	Address string `json:"address"`
	Equiv   uint64 `json:"pusd"`
}

func (s *APIServer) getGlobalRichList(_ context.Context, data json.RawMessage) interface{} {
	params := ParamsGetGlobalRichList{}
	_, _, err := validate(data, &params)
	if err != nil {
		return err
	}

	if params.Count == 0 {
		params.Count = 100
	}

	height := s.Node.GetCurrentSync()
	rates, realHeight, err := s.Node.Pegnet.SelectMostRecentRatesBeforeHeight(nil, s.Node.Pegnet.DB, height+1)
	if err != nil {
		return err
	}

	res := make([]ResultGlobalRichList, 0)
	if realHeight == 0 {
		return res
	}

	rich, err := s.Node.Pegnet.SelectAllBalances()
	if err != nil {
		return err
	}

	for _, r := range rich {
		var usd uint64

		for i := fat2.PTicker(1); i < fat2.PTickerMax; i++ {
			if r.Balances[i] == 0 {
				continue
			}
			c, err := conversions.Convert(int64(r.Balances[i]), rates[i], rates[fat2.PTickerUSD])
			if err != nil {
				return err
			}

			usd += uint64(c)
		}

		if usd == 0 {
			continue
		}

		var entry ResultGlobalRichList
		entry.Address = r.Address.String()
		entry.Equiv = usd

		res = append(res, entry)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Equiv > res[j].Equiv
	})

	if len(res) > params.Count {
		res = res[:params.Count]
	}

	return res
}

type ResultGetRichList struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
	Equiv   uint64 `json:"pusd"`
}

func (s *APIServer) getRichList(_ context.Context, data json.RawMessage) interface{} {
	params := ParamsGetRichList{}
	_, _, err := validate(data, &params)
	if err != nil {
		return err
	}

	if params.Count == 0 {
		params.Count = 100
	}

	height := s.Node.GetCurrentSync()
	rates, rateHeight, err := s.Node.Pegnet.SelectMostRecentRatesBeforeHeight(nil, s.Node.Pegnet.DB, height+1)
	if err != nil {
		return err
	}

	ticker := fat2.StringToTicker(params.Asset) // already validated

	rich, err := s.Node.Pegnet.SelectRichList(ticker, params.Count)
	if err != nil {
		return err
	}

	res := make([]ResultGetRichList, 0)
	for _, r := range rich {
		var entry ResultGetRichList
		entry.Address = r.Address.String()
		entry.Amount = r.Balance
		if rateHeight > 0 {
			c, err := conversions.Convert(int64(r.Balance), rates[ticker], rates[fat2.PTickerUSD])
			if err != nil {
				return err
			}
			entry.Equiv = uint64(c)
		}

		res = append(res, entry)
	}

	return res
}

type ResultGetTransactionStatus struct {
	Height   uint32 `json:"height"`
	Executed uint32 `json:"executed"`
}

func (s *APIServer) getTransactionStatus(_ context.Context, data json.RawMessage) interface{} {
	params := ParamsGetPegnetTransactionStatus{}
	_, _, err := validate(data, &params)
	if err != nil {
		return err
	}

	height, executed, err := s.Node.Pegnet.SelectTransactionHistoryStatus(params.Hash)
	if err != nil {
		return jrpc.ErrorInvalidParams(err)
	}

	if height == 0 {
		return ErrorTransactionNotFound
	}

	var res ResultGetTransactionStatus
	res.Height = height
	res.Executed = executed

	return res
}

// ResultGetTransactions returns history entries.
// `Actions` contains []pegnet.HistoryTransaction.
// `Count` is the total number of possible transactions
// `NextOffset` returns the offset to use to get the next set of records.
//  0 means no more records available
type ResultGetTransactions struct {
	Actions    interface{} `json:"actions"`
	Count      int         `json:"count"`
	NextOffset int         `json:"nextoffset"`
}

func (s *APIServer) getTransactions(forceTxId bool) func(_ context.Context, data json.RawMessage) interface{} {
	return func(_ context.Context, data json.RawMessage) interface{} {
		params := ParamsGetPegnetTransaction{}
		_, _, err := validate(data, &params)
		if err != nil {
			return err
		}

		if forceTxId && params.TxID == "" {
			return jrpc.ErrorInvalidParams(fmt.Errorf("expect txid param to be populated"))
		}

		// using a separate options struct due to golang's circular import restrictions
		var options pegnet.HistoryQueryOptions
		options.Offset = params.Offset
		options.Desc = params.Desc
		options.Transfer = params.Transfer
		options.Conversion = params.Conversion
		options.Coinbase = params.Coinbase
		options.FCTBurn = params.Burn
		options.Asset = params.Asset

		// Are we searching by txid?
		if params.TxID != "" {
			idx, entryhash, err := pegnet.SplitTxID(params.TxID)
			if err != nil {
				return jrpc.ErrorInvalidParams(err.Error())
			}
			options.TxIndex = idx
			options.UseTxIndex = true
			params.txEntryHash = entryhash
		}

		var actions []pegnet.HistoryTransaction
		var count int

		if params.Hash != "" {
			hash := new(factom.Bytes32)
			_ = hash.UnmarshalText([]byte(params.Hash)) // error checked by params.valid
			actions, count, err = s.Node.Pegnet.SelectTransactionHistoryActionsByHash(hash, options)
		} else if params.Address != "" {
			addr, _ := underlyingFA(params.Address) // verified in param
			actions, count, err = s.Node.Pegnet.SelectTransactionHistoryActionsByAddress(&addr, options)
		} else if params.TxID != "" {
			hash := new(factom.Bytes32)
			_ = hash.UnmarshalText([]byte(params.txEntryHash)) // error checked by params.valid
			actions, count, err = s.Node.Pegnet.SelectTransactionHistoryActionsByTxID(hash, options)
		} else {
			actions, count, err = s.Node.Pegnet.SelectTransactionHistoryActionsByHeight(uint32(params.Height), options)
		}

		if err != nil {
			return jrpc.ErrorInvalidParams(err.Error())
		}

		if len(actions) == 0 {
			return ErrorTransactionNotFound
		}

		var res ResultGetTransactions
		res.Count = count
		if params.Offset+len(actions) < count {
			res.NextOffset = params.Offset + len(actions)
		}
		res.Actions = actions

		return res
	}
}

// TODO: This is incompatible with FAT.
type ResultPegnetTickerMap map[fat2.PTicker]uint64

func (r ResultPegnetTickerMap) MarshalJSON() ([]byte, error) {
	strMap := make(map[string]uint64, len(r))
	for ticker, balance := range r {
		strMap[ticker.String()] = balance
	}
	return json.Marshal(strMap)
}
func (r *ResultPegnetTickerMap) UnmarshalJSON(data []byte) error {
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

func (s *APIServer) getPegnetBalances(_ context.Context, data json.RawMessage) interface{} {
	params := ParamsGetPegnetBalances{}
	if _, _, err := validate(data, &params); err != nil {
		return err
	}
	add, _ := underlyingFA(params.Address)

	bals, err := s.Node.Pegnet.SelectBalances(&add)
	if err == sql.ErrNoRows {
		return ErrorAddressNotFound
	}
	if err != nil {
		panic(err) // This is an internal error
	}
	return ResultPegnetTickerMap(bals)
}

type ResultGetIssuance struct {
	SyncStatus ResultGetSyncStatus   `json:"syncstatus"`
	Issuance   ResultPegnetTickerMap `json:"issuance"`
}

func (s *APIServer) getPegnetIssuance(_ context.Context, data json.RawMessage) interface{} {
	issuance, err := s.Node.Pegnet.SelectIssuances()
	if err == sql.ErrNoRows {
		return ErrorAddressNotFound
	}
	if err != nil {
		panic(err) // This is an internal error
	}

	syncStatus := s.getSyncStatus(context.Background(), nil)
	return ResultGetIssuance{
		SyncStatus: syncStatus.(ResultGetSyncStatus),
		Issuance:   issuance,
	}
}

func (s *APIServer) getPegnetRates(ctx context.Context, data json.RawMessage) interface{} {
	params := ParamsGetPegnetRates{}
	if _, _, err := validate(data, &params); err != nil {
		return err
	}

	if params.Height == 0 {
		synced, err := s.Node.Pegnet.SelectSynced(ctx, s.Node.Pegnet.DB)
		if err != nil {
			return err
		}
		params.Height = uint32(synced.Synced)
	}

	rates, err := s.Node.Pegnet.SelectRates(ctx, params.Height)
	if err == sql.ErrNoRows || rates == nil || len(rates) == 0 {
		return ErrorNotFound
	}
	if err != nil {
		panic(err) // This is an internal error
	}

	// The balance results actually works for rates too
	return ResultPegnetTickerMap(rates)
}

func (s *APIServer) sendTransaction(_ context.Context, data json.RawMessage) interface{} {
	params := ParamsSendTransaction{}
	_, _, err := validate(data, &params)
	if err != nil {
		return err
	}
	// defer put()

	ecPrivateKeyString := s.Config.GetString(config.ECPrivateKey)
	var ecPrivateKey factom.EsAddress
	if err = ecPrivateKey.Set(ecPrivateKeyString); err != nil {
		panic(err) // This is an internal error
	}

	entry := params.Entry()
	entry.ChainID = &node.TransactionChain
	// TODO: attempt to apply
	//txErr, err := attemptApplyFAT2TxBatch(chain, entry)
	//if err != nil {
	//	panic(err)
	//}
	//if txErr != nil {
	//	err := ErrorInvalidTransaction
	//	err.Data = txErr.Error()
	//	return err
	//}

	var txID factom.Bytes32
	if !params.DryRun {
		balance, err := ecPrivateKey.ECAddress().GetBalance(nil, s.Node.FactomClient)
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
		txID, err = entry.ComposeCreate(nil, s.Node.FactomClient, ecPrivateKey)
		if err != nil {
			panic(err)
		}
	}

	return struct {
		ChainID *factom.Bytes32 `json:"chainid"`
		TxID    *factom.Bytes32 `json:"txid,omitempty"`
		Hash    *factom.Bytes32 `json:"entryhash"`
	}{ChainID: entry.ChainID, TxID: &txID, Hash: entry.Hash}
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

func (s *APIServer) getSyncStatus(_ context.Context, data json.RawMessage) interface{} {
	heights := new(factom.Heights)
	err := heights.Get(nil, s.Node.FactomClient)
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
			return nil, nil, jrpc.ErrorInvalidParams(`no "params" accepted`)
		}
		return nil, nil, nil
	}
	if len(data) == 0 {
		return nil, nil, params.IsValid()
	}
	if err := unmarshalStrict(data, params); err != nil {
		return nil, nil, jrpc.ErrorInvalidParams(err.Error())
	}
	if err := params.IsValid(); err != nil {
		return nil, nil, err
	}
	//if params.HasIncludePending() && flag.DisablePending {
	//	return nil, nil, ErrorPendingDisabled
	//}
	chainID := params.ValidChainID()
	if chainID != nil {
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

func (s *APIServer) getStats(_ context.Context, data json.RawMessage) interface{} {
	params := ParamsGetStats{}
	if _, _, err := validate(data, &params); err != nil {
		return err
	}
	stats, err := s.Node.Pegnet.SelectStats(context.Background(), *params.Height)
	if err == sql.ErrNoRows {
		return ErrorNotFound
	}
	if err != nil {
		panic(err)
	}

	// The balance results actually works for rates too
	return stats
}

func (s *APIServer) getRecentPegnetRates(ctx context.Context, data json.RawMessage) interface{} {
	params := ParamsGetPegnetRates{}
	if _, _, err := validate(data, &params); err != nil {
		return err
	}

	if params.Height == 0 {
		synced, err := s.Node.Pegnet.SelectSynced(ctx, s.Node.Pegnet.DB)
		if err != nil {
			return err
		}
		params.Height = uint32(synced.Synced)
	}

	rates, err := s.Node.Pegnet.SelectRecentRates(ctx, params.Height)
	if err == sql.ErrNoRows || rates == nil || len(rates) == 0 {
		return ErrorNotFound
	}
	if err != nil {
		panic(err) // This is an internal error
	}

	// The balance results actually works for rates too
	return rates
}
